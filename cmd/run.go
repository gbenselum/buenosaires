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
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/spf13/cobra"
)

// runCmd implements the main monitoring loop that watches a Git repository
// for new shell scripts and executes them after validation.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the buenosaires monitor",
	Long:  `This command starts the buenosaires monitor, which watches a repository for new shell scripts and executes them.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load the global configuration from ~/.buenosaires/config.toml
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			log.Fatalf("Failed to load global config: %v", err)
		}

		// Load the status file that tracks script execution history
		status, err := status.LoadStatus(".")
		if err != nil {
			log.Fatalf("Failed to load status file: %v", err)
		}

		// Start the web server if enabled
		if globalConfig.GUI.Enabled {
			addr := fmt.Sprintf(":%d", globalConfig.GUI.Port)
			go web.StartServer(addr, globalConfig.LogDir)
		}

		// Open the Git repository in the current directory
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

		// Main monitoring loop - polls the repository every 10 seconds
		for {
			syncInterval := time.Duration(globalConfig.SyncInterval) * time.Second
			if globalConfig.SyncInterval == 0 {
				syncInterval = 180 * time.Second
			}

			// Fetch the latest changes from the remote
			err := repo.Fetch(&git.FetchOptions{})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("Failed to fetch from remote: %v", err)
				time.Sleep(syncInterval)
				continue
			}

			// Get the latest commit on the branch
			branchRef, err := repo.Reference(branchRefName, true)
			if err != nil {
				log.Printf("Failed to get branch reference: %v", err)
				time.Sleep(syncInterval)
				continue
			}

			latestCommitHash := branchRef.Hash()

			// Process new commits if the hash has changed
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
				if repoConfig.Plugins["shell"] {
					for _, change := range changes {
						if isNewShellScript(change) {
							scriptName := change.To.Name
							if s, ok := status.Scripts[scriptName]; ok && s.OverallStatus == "success" {
								log.Printf("Script %s already processed successfully, skipping.", scriptName)
								continue
							}

							log.Printf("New shell script found: %s", scriptName)
							// Initialize the script status as pending
							status.UpdateScriptStatus(scriptName, "pending", "skipped", "pending", "pending")
							status.SaveStatus(".")

							// Retrieve the file content from the Git tree
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

							// Create a temporary file to store the script for validation and execution
							tmpfile, err := os.CreateTemp("", "script-*.sh")
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

							// Validate the script using shellcheck and syntax checking
							plugin := shell.ShellPlugin{}
							lintOutput, err := plugin.LintAndValidate(tmpfile.Name())
							lintPassed := err == nil
							if err != nil {
								log.Printf("Script validation failed for %s: %v\n%s", scriptName, err, lintOutput)
								status.UpdateScriptStatus(scriptName, "failure", "skipped", "pending", "failure")
								status.SaveStatus(".")
								plugin.UpdateAssetAfterRun(scriptName, repoConfig.User, latestCommitHash.String(), lintOutput, lintPassed, 0, "failure")
								continue // Skip execution of invalid scripts
							}
							log.Printf("Script validation successful for %s:\n%s", scriptName, lintOutput)
							status.UpdateScriptStatus(scriptName, "success", "skipped", "pending", "pending")
							status.SaveStatus(".")

							// Execute the script
							startTime := time.Now()
							execOutput, err := plugin.Run(tmpfile.Name(), repoConfig.AllowSudo)
							runDuration := time.Since(startTime)
							runStatus := "success"
							if err != nil {
								log.Printf("Failed to execute script %s: %v", scriptName, err)
								status.UpdateScriptStatus(scriptName, "success", "skipped", "failure", "failure")
								status.SaveStatus(".")
								runStatus = "failure"
							} else {
								status.UpdateScriptStatus(scriptName, "success", "skipped", "success", "success")
								status.SaveStatus(".")
							}
							plugin.UpdateAssetAfterRun(scriptName, repoConfig.User, latestCommitHash.String(), execOutput, lintPassed, runDuration, runStatus)

							// Write the combined lint and execution output to a log file
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
								err := os.WriteFile(logFile, []byte(logContent), 0644)
								if err != nil {
									log.Printf("Failed to write log file: %v", err)
								}
							}
						}
					}
				}

			lastCommitHash = latestCommitHash
		}

		// Wait before polling again
		if globalConfig.SyncInterval == 0 {
			time.Sleep(180 * time.Second)
		} else {
			time.Sleep(time.Duration(globalConfig.SyncInterval) * time.Second)
		}
		}
	},
}

// isNewShellScript checks if a Git change represents a newly added shell script.
// It returns true only if the change is an insert operation and the file has a .sh extension.
func isNewShellScript(change *object.Change) bool {
	action, err := change.Action()
	if err != nil {
		return false
	}
	return action == merkletrie.Insert && strings.HasSuffix(change.To.Name, ".sh")
}

// init registers the run command with the root command.
func init() {
	rootCmd.AddCommand(runCmd)
}