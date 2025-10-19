// Package cmd provides the command-line interface for Buenos Aires.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"buenosaires/internal/config"

	"github.com/spf13/cobra"
)

// installCmd handles the interactive installation and configuration process.
// It prompts the user for configuration values and saves them to the global config file.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and configure buenosaires",
	Long:  `This command installs and configures the buenosaires tool, setting up the necessary configuration file in your home directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		// Prompt for username - this will be the default user for running scripts
		fmt.Print("Enter your username: ")
		user, _ := reader.ReadString('\n')
		user = strings.TrimSpace(user)

		// Prompt for log directory - where script execution logs will be saved
		fmt.Print("Enter the folder to save logs: ")
		logDir, _ := reader.ReadString('\n')
		logDir = strings.TrimSpace(logDir)

		// Prompt for branch to monitor - typically "main" or "master"
		fmt.Print("Enter the branch to monitor (e.g., main): ")
		branch, _ := reader.ReadString('\n')
		branch = strings.TrimSpace(branch)

		// Prompt for the repository to scan
		fmt.Print("Enter the repository to scan (default: https://github.com/gbenselum/buenosaires_test): ")
		repoURL, _ := reader.ReadString('\n')
		repoURL = strings.TrimSpace(repoURL)
		if repoURL == "" {
			repoURL = "https://github.com/gbenselum/buenosaires_test"
		}

		// Prompt for Web GUI configuration
		fmt.Print("Enable Web GUI? (y/n): ")
		enableGUIStr, _ := reader.ReadString('\n')
		enableGUI := strings.TrimSpace(strings.ToLower(enableGUIStr)) == "y"

		// If Web GUI is enabled, prompt for port number
		var port int
		if enableGUI {
			fmt.Print("Enter the port for the Web GUI (e.g., 9099): ")
			portStr, _ := reader.ReadString('\n')
			portStr = strings.TrimSpace(portStr)
			if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
				fmt.Println("Invalid port number, defaulting to 9099")
				port = 9099
			}
		}

		// Create the global configuration object
		cfg := config.GlobalConfig{
			User:          user,
			LogDir:        logDir,
			Branch:        branch,
			RepositoryURL: repoURL,
			GUI: config.GUIConfig{
				Enabled: enableGUI,
				Port:    port,
			},
		}

		// Save the configuration to the global config file
		if err := config.SaveGlobalConfig(cfg); err != nil {
			fmt.Println("Error saving configuration:", err)
			return
		}

		fmt.Println("Configuration saved successfully!")
	},
}

// init registers the install command with the root command.
func init() {
	rootCmd.AddCommand(installCmd)
}