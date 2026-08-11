// Package cmd provides the command-line interface for Buenos Aires.
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"buenosaires/internal/config"
	"buenosaires/internal/status"
	"buenosaires/internal/web"
	"buenosaires/plugins/shell"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/spf13/cobra"
)

// runCmd implements the main monitoring loop that watches a Git repository
// for shell scripts and executes them after validation.
//
// Design notes:
//   - Change detection is driven by the remote tracking ref after a fetch,
//     so it works even when the local worktree is dirty or a pull fails.
//   - The last processed commit hash is persisted in status.json, which means
//     the monitor resumes incrementally across restarts and performs an
//     initial sync of existing scripts on first run (instead of doing nothing).
//   - Both inserted AND modified scripts are processed. A fix or a new
//     version of a script therefore re-deploys, and a script that previously
//     failed can be retried.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the buenosaires monitor",
	Long:  `This command starts the buenosaires monitor, which watches a repository for shell scripts and executes them.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load the global configuration from ~/.buenosaires/config.toml
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			log.Fatalf("Failed to load global config: %v", err)
		}

		// Load the status file that tracks script execution history.
		// The persisted LastCommit hash makes change detection resume
		// incrementally across restarts.
		st, err := status.LoadStatus(".")
		if err != nil {
			log.Fatalf("Failed to load status file: %v", err)
		}

		// Determine the effective log directory, honoring repo overrides.
		logDir := globalConfig.LogDir
		if rc, err := config.LoadRepoConfig("."); err == nil && rc.LogDir != "" {
			logDir = rc.LogDir
		}

		// Start the web server if enabled. Failures are logged, not fatal,
		// so a web server error never kills the monitor.
		if globalConfig.GUI.Enabled {
			web.StartServer(fmt.Sprintf(":%d", globalConfig.GUI.Port), logDir)
		}

		repo, err := openOrCloneRepo(globalConfig.RepositoryURL)
		if err != nil {
			log.Fatalf("Failed to open repository: %v", err)
		}

		// Resolve the remote tracking ref so change detection is independent
		// of the local worktree state.
		remoteRefName := plumbing.NewRemoteReferenceName("origin", globalConfig.Branch)

		syncInterval := getSyncInterval(globalConfig)
		log.Printf("Starting to monitor branch '%s' every %s", globalConfig.Branch, syncInterval)

		for {
			// Fetch latest changes. Failure is non-fatal: we keep monitoring
			// against the last known remote ref.
			if err := repo.Fetch(&git.FetchOptions{RemoteName: "origin"}); err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("Failed to fetch from remote: %v", err)
			}

			remoteRef, err := repo.Reference(remoteRefName, true)
			if err != nil {
				log.Printf("Failed to resolve remote ref %s: %v", remoteRefName, err)
				time.Sleep(syncInterval)
				continue
			}
			latestCommitHash := remoteRef.Hash()

			if latestCommitHash.String() != st.LastCommit {
				log.Printf("New commits detected: %s", latestCommitHash.String())

				repoConfig, err := config.LoadRepoConfig(".")
				if err != nil {
					log.Printf("Failed to load repo config: %v", err)
				}

				// Sudo requires BOTH the host operator's global opt-in and the
				// repository's allow_sudo flag. A malicious repository cannot
				// escalate privileges on its own.
				allowSudo := globalConfig.AllowSudo && repoConfig.AllowSudo

				if pluginConfig, ok := repoConfig.Plugins["shell"]; ok && pluginConfig.Enabled {
					folderToScan := pluginConfig.FolderToScan
					if folderToScan == "" {
						folderToScan = "shell"
					}
					if err := processChanges(repo, st, latestCommitHash, folderToScan, repoConfig, globalConfig, allowSudo); err != nil {
						log.Printf("Failed to process changes: %v", err)
					}
				}

				st.LastCommit = latestCommitHash.String()
				if err := st.SaveStatus("."); err != nil {
					log.Printf("Failed to save status: %v", err)
				}
			}

			// Best-effort worktree update. The monitor does not depend on it:
			// change detection and execution read from git objects, and
			// scripts are materialized from the tree before execution.
			if w, err := repo.Worktree(); err == nil {
				if err := w.Pull(&git.PullOptions{RemoteName: "origin"}); err != nil && err != git.NoErrAlreadyUpToDate {
					log.Printf("Worktree pull failed (monitoring continues from remote refs): %v", err)
				}
			}

			// Wait before polling again
			time.Sleep(syncInterval)
		}
	},
}

// openOrCloneRepo opens the Git repository in the current directory, cloning
// the configured repository if no repository exists yet.
func openOrCloneRepo(repositoryURL string) (*git.Repository, error) {
	repo, err := git.PlainOpen(".")
	if err == nil {
		return repo, nil
	}
	if err != git.ErrRepositoryNotExists {
		return nil, err
	}

	log.Printf("Cloning repository from %s", repositoryURL)
	repo, err = git.PlainClone(".", false, &git.CloneOptions{URL: repositoryURL})
	if err != nil {
		return nil, fmt.Errorf("clone failed: %w (the workspace must be an empty directory or an existing clone of %s)", err, repositoryURL)
	}
	return repo, nil
}

// processChanges diffs the previously processed commit against the latest
// commit and processes every relevant shell script change. When no commit has
// been processed yet (or the previous one is no longer available), it performs
// an initial sync over all existing scripts in the scanned folder.
func processChanges(repo *git.Repository, st *status.Status, latestCommitHash plumbing.Hash, folderToScan string, repoConfig config.RepoConfig, globalConfig config.GlobalConfig, allowSudo bool) error {
	latestCommit, err := repo.CommitObject(latestCommitHash)
	if err != nil {
		return fmt.Errorf("failed to get latest commit object: %w", err)
	}
	latestTree, err := latestCommit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get latest commit tree: %w", err)
	}

	var changes []*object.Change
	if st.LastCommit == "" {
		log.Printf("Initial sync: processing existing scripts under %s/", folderToScan)
		if changes, err = initialSyncChanges(latestTree, folderToScan); err != nil {
			return fmt.Errorf("failed to walk initial tree: %w", err)
		}
	} else {
		baseCommit, err := repo.CommitObject(plumbing.NewHash(st.LastCommit))
		if err != nil {
			// The base commit is no longer available (e.g. shallow clone or
			// garbage collection): fall back to an initial sync instead of
			// failing the entire run.
			log.Printf("Last processed commit %s not found (%v); falling back to initial sync", st.LastCommit, err)
			if changes, err = initialSyncChanges(latestTree, folderToScan); err != nil {
				return fmt.Errorf("failed to walk initial tree: %w", err)
			}
		} else {
			baseTree, err := baseCommit.Tree()
			if err != nil {
				return fmt.Errorf("failed to get base commit tree: %w", err)
			}
			changes, err = object.DiffTree(baseTree, latestTree)
			if err != nil {
				return fmt.Errorf("failed to diff trees: %w", err)
			}
		}
	}

	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			log.Printf("Failed to determine change action: %v", err)
			continue
		}

		// Resolve the script name from the appropriate side of the change.
		var scriptName string
		if action == merkletrie.Delete && change.From.Name != "" {
			scriptName = change.From.Name
		} else if change.To.Name != "" {
			scriptName = change.To.Name
		}
		if scriptName == "" {
			continue
		}

		if action == merkletrie.Delete {
			if isShellScriptInFolder(scriptName, folderToScan) {
				log.Printf("Script %s removed from repository, cleaning up status.", scriptName)
				delete(st.Scripts, scriptName)
			}
			continue
		}

		if !shouldProcessScript(action, scriptName, folderToScan, st.Scripts[scriptName].OverallStatus) {
			continue
		}

		log.Printf("%s detected for %s", action, scriptName)
		processScript(scriptName, latestTree, latestCommitHash, st, repoConfig, globalConfig, allowSudo)
	}
	return nil
}

// initialSyncChanges returns a synthetic list of Insert changes for every
// shell script present in the given tree under the scanned folder.
func initialSyncChanges(tree *object.Tree, folderToScan string) ([]*object.Change, error) {
	var changes []*object.Change
	err := tree.Files().ForEach(func(f *object.File) error {
		if isShellScriptInFolder(f.Name, folderToScan) {
			changes = append(changes, &object.Change{
				To: object.ChangeEntry{Name: f.Name},
			})
		}
		return nil
	})
	return changes, err
}

// isShellScriptInFolder reports whether the given repository path is a shell
// script located under the folder being scanned.
func isShellScriptInFolder(scriptName, folderToScan string) bool {
	return strings.HasSuffix(scriptName, ".sh") && strings.HasPrefix(scriptName, folderToScan+"/")
}

// shouldProcessScript decides whether a changed script should be processed.
//   - Insert: process unless the script was already processed successfully
//     (avoids re-running on restarts and initial syncs).
//   - Modify: always process, so fixes and new versions re-deploy.
//   - Delete and other actions: never process (deletes are cleaned up
//     separately by the caller).
func shouldProcessScript(action merkletrie.Action, scriptName, folderToScan, prevOverallStatus string) bool {
	if !isShellScriptInFolder(scriptName, folderToScan) {
		return false
	}
	switch action {
	case merkletrie.Modify:
		return true
	case merkletrie.Insert:
		return prevOverallStatus != status.StatusSuccess
	default:
		return false
	}
}

// getSyncInterval returns the configured sync interval, defaulting to 180 seconds if not set.
func getSyncInterval(config config.GlobalConfig) time.Duration {
	if config.SyncInterval == 0 {
		return 180 * time.Second
	}
	return time.Duration(config.SyncInterval) * time.Second
}

// processScript handles the complete lifecycle of a shell script: validation,
// execution, logging, and status/asset bookkeeping.
func processScript(scriptName string, latestTree *object.Tree, latestCommitHash plumbing.Hash, st *status.Status, repoConfig config.RepoConfig, globalConfig config.GlobalConfig, allowSudo bool) {
	// Defense in depth: reject unsafe paths even though git trees cannot
	// normally contain traversal components.
	if clean := filepath.Clean(scriptName); clean != scriptName || strings.Contains(scriptName, "..") {
		log.Printf("Rejecting script with unsafe path: %s", scriptName)
		return
	}

	// Initialize the script status as pending
	st.UpdateScriptStatus(scriptName, status.StatusPending, status.StatusSkipped, status.StatusPending, status.StatusPending)
	if err := st.SaveStatus("."); err != nil {
		log.Printf("Failed to save status: %v", err)
	}

	// Retrieve the file content from the Git tree
	file, err := latestTree.File(scriptName)
	if err != nil {
		log.Printf("Failed to get file from tree: %v", err)
		return
	}
	content, err := file.Contents()
	if err != nil {
		log.Printf("Failed to get file contents: %v", err)
		return
	}

	// Materialize the script at its real path so that relative paths inside
	// the script resolve against the repository layout, matching how a
	// developer would run it locally. Executing from a /tmp scratch file
	// would break scripts that reference sibling files.
	// The file is written with its committed git mode so the worktree stays
	// clean. Execution always goes through `bash <path>` with no shell
	// interpolation, so executable bits are not required.
	localPath := filepath.FromSlash(scriptName)
	if dir := filepath.Dir(localPath); dir != "." {
		// #nosec G301 -- directory permissions mirror standard repository dirs.
		if err := os.MkdirAll(dir, 0o750); err != nil {
			log.Printf("Failed to create script directory %s: %v", dir, err)
			return
		}
	}
	mode := os.FileMode(0o644)
	if file.Mode == filemode.Executable {
		mode = 0o755
	}
	// #nosec G306 -- mode mirrors the file's committed git mode (see above).
	if err := os.WriteFile(localPath, []byte(content), mode); err != nil {
		log.Printf("Failed to materialize script %s: %v", localPath, err)
		return
	}

	// Validate the script using shellcheck and syntax checking
	plugin := shell.ShellPlugin{}
	lintOutput, err := plugin.LintAndValidate(localPath)
	lintPassed := err == nil
	if err != nil {
		log.Printf("Script validation failed for %s: %v\n%s", scriptName, err, lintOutput)
		st.UpdateScriptStatus(scriptName, status.StatusFailure, status.StatusSkipped, status.StatusPending, status.StatusFailure)
		if err := st.SaveStatus("."); err != nil {
			log.Printf("Failed to save status: %v", err)
		}
		if err := plugin.UpdateAssetAfterRun(scriptName, repoConfig.User, latestCommitHash.String(), lintOutput, lintPassed, 0, status.StatusFailure); err != nil {
			log.Printf("Failed to update asset after run: %v", err)
		}
		return // Skip execution of invalid scripts
	}
	log.Printf("Script validation successful for %s:\n%s", scriptName, lintOutput)
	st.UpdateScriptStatus(scriptName, status.StatusSuccess, status.StatusSkipped, status.StatusPending, status.StatusPending)
	if err := st.SaveStatus("."); err != nil {
		log.Printf("Failed to save status: %v", err)
	}

	// Execute the script
	startTime := time.Now()
	execOutput, err := plugin.Run(localPath, allowSudo)
	runDuration := time.Since(startTime)
	runStatus := status.StatusSuccess
	if err != nil {
		log.Printf("Failed to execute script %s: %v", scriptName, err)
		runStatus = status.StatusFailure
	}
	st.UpdateScriptStatus(scriptName, status.StatusSuccess, status.StatusSkipped, runStatus, runStatus)
	if err := st.SaveStatus("."); err != nil {
		log.Printf("Failed to save status: %v", err)
	}
	if err := plugin.UpdateAssetAfterRun(scriptName, repoConfig.User, latestCommitHash.String(), execOutput, lintPassed, runDuration, runStatus); err != nil {
		log.Printf("Failed to update asset after run: %v", err)
	}

	// Write the combined lint and execution output to a log file
	logDir := repoConfig.LogDir
	if logDir == "" {
		logDir = globalConfig.LogDir
	}
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0750); err != nil {
			log.Printf("Failed to create log directory: %v", err)
		} else {
			logFile := filepath.Join(logDir, filepath.Base(scriptName)+".log")
			logContent := fmt.Sprintf("--- LINT OUTPUT ---\n%s\n--- EXECUTION OUTPUT ---\n%s", lintOutput, execOutput)
			if err := os.WriteFile(logFile, []byte(logContent), 0600); err != nil {
				log.Printf("Failed to write log file: %v", err)
			}
		}
	}
}

// init registers the run command with the root command.
func init() {
	rootCmd.AddCommand(runCmd)
}
