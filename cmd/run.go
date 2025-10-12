package cmd

import (
	"fmt"
	"io/ioutil"
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
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the buenosaires monitor",
	Long:  `This command starts the buenosaires monitor, which watches a repository for new shell scripts and executes them.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load global config
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			log.Fatalf("Failed to load global config: %v", err)
		}

		// Load status file
		status, err := status.LoadStatus(".")
		if err != nil {
			log.Fatalf("Failed to load status file: %v", err)
		}

		// Start the web server if enabled
		if globalConfig.GUI.Enabled {
			addr := fmt.Sprintf(":%d", globalConfig.GUI.Port)
			go web.StartServer(addr, globalConfig.LogDir)
		}

		// Open the current repository
		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Failed to open repository: %v", err)
		}

		// Get the reference for the branch to monitor
		branchRefName := plumbing.NewBranchReferenceName(globalConfig.Branch)
		branchRef, err := repo.Reference(branchRefName, true)
		if err != nil {
			log.Fatalf("Failed to get branch reference: %v", err)
		}

		var lastCommitHash plumbing.Hash
		if branchRef != nil {
			lastCommitHash = branchRef.Hash()
		}

		log.Printf("Starting to monitor branch '%s'", globalConfig.Branch)

		for {
			// Fetch the latest changes from the remote
			err := repo.Fetch(&git.FetchOptions{})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("Failed to fetch from remote: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			// Get the latest commit on the branch
			branchRef, err := repo.Reference(branchRefName, true)
			if err != nil {
				log.Printf("Failed to get branch reference: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			latestCommitHash := branchRef.Hash()

			// If the commit hash has changed, check for new scripts
			if latestCommitHash != lastCommitHash {
				log.Printf("New commit detected: %s", latestCommitHash.String())

				// Get the commit objects
				latestCommit, err := repo.CommitObject(latestCommitHash)
				if err != nil {
					log.Printf("Failed to get latest commit object: %v", err)
					lastCommitHash = latestCommitHash
					continue
				}

				var lastCommit *object.Commit
				if lastCommitHash != (plumbing.Hash{}) {
					lastCommit, err = repo.CommitObject(lastCommitHash)
					if err != nil {
						log.Printf("Failed to get last commit object: %v", err)
						lastCommitHash = latestCommitHash
						continue
					}
				}

				// Get the trees for both commits
				latestTree, err := latestCommit.Tree()
				if err != nil {
					log.Printf("Failed to get latest commit tree: %v", err)
					lastCommitHash = latestCommitHash
					continue
				}

				var lastTree *object.Tree
				if lastCommit != nil {
					lastTree, err = lastCommit.Tree()
					if err != nil {
						log.Printf("Failed to get last commit tree: %v", err)
						lastCommitHash = latestCommitHash
						continue
					}
				}

				// Compare the trees to find new files
				changes, err := object.DiffTree(lastTree, latestTree)
				if err != nil {
					log.Printf("Failed to diff trees: %v", err)
					lastCommitHash = latestCommitHash
					continue
				}

				// Load repo-specific config
				repoConfig, err := config.LoadRepoConfig(".")
				if err != nil {
					log.Printf("Failed to load repo config: %v", err)
				}

				// Check for new .sh files
				for _, change := range changes {
					if isNewShellScript(change) {
						scriptName := change.To.Name
						if s, ok := status.Scripts[scriptName]; ok && s.OverallStatus == "success" {
							log.Printf("Script %s already processed successfully, skipping.", scriptName)
							continue
						}

						log.Printf("New shell script found: %s", scriptName)
						status.UpdateScriptStatus(scriptName, "pending", "skipped", "pending", "pending")
						status.SaveStatus(".")

						// Get the file content
						file, err := latestTree.File(scriptName)
						if err != nil {
							log.Printf("Failed to get file from tree: %v", err)
							continue
						}
						content, err := file.Contents()
						if err != nil {
							log.Printf("Failed to get file contents: %v", err)
							continue
						}

						// Create a temporary file to execute
						tmpfile, err := ioutil.TempFile("", "script-*.sh")
						if err != nil {
							log.Printf("Failed to create temporary file: %v", err)
							continue
						}
						defer os.Remove(tmpfile.Name())

						if _, err := tmpfile.Write([]byte(content)); err != nil {
							log.Printf("Failed to write to temporary file: %v", err)
							tmpfile.Close()
							continue
						}
						tmpfile.Close()

						// Lint and validate the script
						plugin := shell.ShellPlugin{}
						lintOutput, err := plugin.LintAndValidate(tmpfile.Name())
						if err != nil {
							log.Printf("Script validation failed for %s: %v\n%s", scriptName, err, lintOutput)
							status.UpdateScriptStatus(scriptName, "failure", "skipped", "pending", "failure")
							status.SaveStatus(".")
							continue // Skip invalid scripts
						}
						log.Printf("Script validation successful for %s:\n%s", scriptName, lintOutput)
						status.UpdateScriptStatus(scriptName, "success", "skipped", "pending", "pending")
						status.SaveStatus(".")

						// Execute the script
						execOutput, err := plugin.Run(tmpfile.Name(), repoConfig.AllowSudo)
						if err != nil {
							log.Printf("Failed to execute script %s: %v", scriptName, err)
							status.UpdateScriptStatus(scriptName, "success", "skipped", "failure", "failure")
							status.SaveStatus(".")
						} else {
							status.UpdateScriptStatus(scriptName, "success", "skipped", "success", "success")
							status.SaveStatus(".")
						}

						// Log the output
						logDir := repoConfig.LogDir
						if logDir == "" {
							logDir = globalConfig.LogDir
						}
						if logDir != "" {
							if _, err := os.Stat(logDir); os.IsNotExist(err) {
								os.MkdirAll(logDir, 0755)
							}
							logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", filepath.Base(scriptName)))
							logContent := fmt.Sprintf("--- LINT OUTPUT ---\n%s\n--- EXECUTION OUTPUT ---\n%s", lintOutput, execOutput)
							err := ioutil.WriteFile(logFile, []byte(logContent), 0644)
							if err != nil {
								log.Printf("Failed to write log file: %v", err)
							}
						}
					}
				}

				lastCommitHash = latestCommitHash
			}

			time.Sleep(10 * time.Second) // Poll every 10 seconds
		}
	},
}

func isNewShellScript(change *object.Change) bool {
	action, err := change.Action()
	if err != nil {
		return false
	}
	return action == merkletrie.Insert && strings.HasSuffix(change.To.Name, ".sh")
}

func init() {
	rootCmd.AddCommand(runCmd)
}