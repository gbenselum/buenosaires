package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"buenosaires/internal/config"
	"buenosaires/internal/status"
	"buenosaires/internal/web"
	"buenosaires/plugins/docker"
	"buenosaires/plugins/shell"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/spf13/cobra"
)

const (
	// Default polling interval for checking repository updates
	DefaultPollInterval = 10 * time.Second
	// Maximum size for shell scripts (10MB)
	MaxScriptSize = 10 * 1024 * 1024
	// Default script execution timeout
	DefaultScriptTimeout = 5 * time.Minute
	// File permissions
	DirPerm  = 0755
	FilePerm = 0644
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
						// Use thread-safe getter
						if s, ok := status.GetScriptStatus(scriptName); ok && s.OverallStatus == "success" {
							log.Printf("Script %s already processed successfully, skipping.", scriptName)
							continue
						}

						log.Printf("New shell script found: %s", scriptName)
						status.UpdateScriptStatus(scriptName, "pending", "skipped", "pending", "pending")
						if err := status.SaveStatus("."); err != nil {
							log.Printf("Failed to save status: %v", err)
						}

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

						// Security: Check script size
						if len(content) > MaxScriptSize {
							log.Printf("Script %s exceeds maximum size of %d bytes", scriptName, MaxScriptSize)
							status.UpdateScriptStatus(scriptName, "failure", "skipped", "pending", "failure")
							if err := status.SaveStatus("."); err != nil {
								log.Printf("Failed to save status: %v", err)
							}
							continue
						}

						// Create a temporary file to execute
						tmpfile, err := os.CreateTemp("", "script-*.sh")
						if err != nil {
							log.Printf("Failed to create temporary file: %v", err)
							continue
						}
						tmpfileName := tmpfile.Name()

						if _, err := tmpfile.Write([]byte(content)); err != nil {
							log.Printf("Failed to write to temporary file: %v", err)
							tmpfile.Close()
							os.Remove(tmpfileName)
							continue
						}
						if err := tmpfile.Close(); err != nil {
							log.Printf("Failed to close temporary file: %v", err)
							os.Remove(tmpfileName)
							continue
						}

						// Lint and validate the script
						plugin := shell.ShellPlugin{}
						
						// Create context with timeout for validation
						ctx, cancel := context.WithTimeout(context.Background(), DefaultScriptTimeout)
						defer cancel()
						
						lintOutput, err := plugin.LintAndValidate(tmpfileName)
						if err != nil {
							log.Printf("Script validation failed for %s: %v\n%s", scriptName, err, lintOutput)
							status.UpdateScriptStatus(scriptName, "failure", "skipped", "pending", "failure")
							if err := status.SaveStatus("."); err != nil {
								log.Printf("Failed to save status: %v", err)
							}
							os.Remove(tmpfileName)
							continue // Skip invalid scripts
						}
						log.Printf("Script validation successful for %s:\n%s", scriptName, lintOutput)
						status.UpdateScriptStatus(scriptName, "success", "skipped", "pending", "pending")
						if err := status.SaveStatus("."); err != nil {
							log.Printf("Failed to save status: %v", err)
						}

						// Security: Log sudo execution for audit
						if repoConfig.AllowSudo {
							log.Printf("WARNING: Executing script %s with sudo privileges", scriptName)
						}

						// Execute the script with timeout
						execOutput, err := plugin.Run(tmpfileName, repoConfig.AllowSudo)
						_ = ctx // Use context (prepared for future timeout implementation in plugin)
						
						if err != nil {
							log.Printf("Failed to execute script %s: %v", scriptName, err)
							status.UpdateScriptStatus(scriptName, "success", "skipped", "failure", "failure")
							if err := status.SaveStatus("."); err != nil {
								log.Printf("Failed to save status: %v", err)
							}
						} else {
							status.UpdateScriptStatus(scriptName, "success", "skipped", "success", "success")
							if err := status.SaveStatus("."); err != nil {
								log.Printf("Failed to save status: %v", err)
							}
						}

						// Clean up temporary file
						os.Remove(tmpfileName)

						// Log the output
						logDir := repoConfig.LogDir
						if logDir == "" {
							logDir = globalConfig.LogDir
						}
						if logDir != "" {
							if _, err := os.Stat(logDir); os.IsNotExist(err) {
								if err := os.MkdirAll(logDir, DirPerm); err != nil {
									log.Printf("Failed to create log directory: %v", err)
								}
							}
							logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", filepath.Base(scriptName)))
							logContent := fmt.Sprintf("--- LINT OUTPUT ---\n%s\n--- EXECUTION OUTPUT ---\n%s", lintOutput, execOutput)
							if err := os.WriteFile(logFile, []byte(logContent), FilePerm); err != nil {
								log.Printf("Failed to write log file: %v", err)
							}
						}
					}
					
					// Process Docker container files if Docker plugin is enabled
					if repoConfig.Docker.Enabled && globalConfig.Plugins["docker"] {
						if isNewOrModifiedContainerFile(change) {
							containerPath := change.To.Name
							containerDir := filepath.Dir(containerPath)
							containerName := filepath.Base(containerDir)
							
							// Use container directory name as identifier
							statusKey := fmt.Sprintf("container:%s", containerName)
							
							if s, ok := status.GetScriptStatus(statusKey); ok && s.OverallStatus == "success" {
								log.Printf("Container %s already processed successfully, skipping.", containerName)
								continue
							}

							log.Printf("New/modified container file found: %s", containerPath)
							status.UpdateScriptStatus(statusKey, "pending", "skipped", "pending", "pending")
							if err := status.SaveStatus("."); err != nil {
								log.Printf("Failed to save status: %v", err)
							}

						// For Docker builds, we need the entire directory context
						// Verify the file exists in the tree
						_, err := latestTree.File(containerPath)
						if err != nil {
							log.Printf("Failed to get container file from tree: %v", err)
							continue
						}
						
						dockerPlugin := docker.DockerPlugin{}
							
							// Determine the full path to the container file
							fullContainerPath := filepath.Join(".", containerPath)
							
							// Lint and validate the Dockerfile/Containerfile
							lintOutput, err := dockerPlugin.LintAndValidate(fullContainerPath)
							if err != nil {
								log.Printf("Container validation failed for %s: %v\n%s", containerName, err, lintOutput)
								status.UpdateScriptStatus(statusKey, "failure", "skipped", "pending", "failure")
								if err := status.SaveStatus("."); err != nil {
									log.Printf("Failed to save status: %v", err)
								}
								continue
							}
							log.Printf("Container validation successful for %s:\n%s", containerName, lintOutput)
							status.UpdateScriptStatus(statusKey, "success", "skipped", "pending", "pending")
							if err := status.SaveStatus("."); err != nil {
								log.Printf("Failed to save status: %v", err)
							}

							// Build (and optionally run) the container
							imageTag := repoConfig.Docker.DefaultTag
							if imageTag == "" {
								imageTag = "latest"
							}
							
							imageName := containerName
							if repoConfig.Docker.ImagePrefix != "" {
								imageName = repoConfig.Docker.ImagePrefix + imageName
							}
							
							log.Printf("Building Docker image: %s:%s", imageName, imageTag)
							if repoConfig.Docker.AutoRun {
								log.Printf("WARNING: auto_run is enabled, container will be started automatically")
							}
							
							execOutput, err := dockerPlugin.Run(fullContainerPath, imageName, imageTag, repoConfig.Docker.AutoRun)
							if err != nil {
								log.Printf("Failed to build container %s: %v", containerName, err)
								status.UpdateScriptStatus(statusKey, "success", "skipped", "failure", "failure")
								if err := status.SaveStatus("."); err != nil {
									log.Printf("Failed to save status: %v", err)
								}
							} else {
								status.UpdateScriptStatus(statusKey, "success", "skipped", "success", "success")
								if err := status.SaveStatus("."); err != nil {
									log.Printf("Failed to save status: %v", err)
								}
							}

							// Log the output
							logDir := repoConfig.LogDir
							if logDir == "" {
								logDir = globalConfig.LogDir
							}
							if logDir != "" {
								if _, err := os.Stat(logDir); os.IsNotExist(err) {
									if err := os.MkdirAll(logDir, DirPerm); err != nil {
										log.Printf("Failed to create log directory: %v", err)
									}
								}
								logFile := filepath.Join(logDir, fmt.Sprintf("%s-container.log", containerName))
								logContent := fmt.Sprintf("--- LINT OUTPUT ---\n%s\n--- BUILD/RUN OUTPUT ---\n%s", lintOutput, execOutput)
								if err := os.WriteFile(logFile, []byte(logContent), FilePerm); err != nil {
									log.Printf("Failed to write log file: %v", err)
								}
							}
						}
					}
				}

				lastCommitHash = latestCommitHash
			}

			time.Sleep(DefaultPollInterval)
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

func isNewOrModifiedContainerFile(change *object.Change) bool {
	action, err := change.Action()
	if err != nil {
		return false
	}
	
	// Check if it's in the Containers folder and is a Dockerfile or Containerfile
	path := change.To.Name
	if !strings.HasPrefix(path, "Containers/") && !strings.HasPrefix(path, "containers/") {
		return false
	}
	
	filename := filepath.Base(path)
	isContainerFile := filename == "Dockerfile" || filename == "Containerfile"
	
	// Accept both new files and modifications
	return (action == merkletrie.Insert || action == merkletrie.Modify) && isContainerFile
}

func init() {
	rootCmd.AddCommand(runCmd)
}