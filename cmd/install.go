package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"buenosaires/internal/config"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and configure buenosaires",
	Long:  `This command installs and configures the buenosaires tool, setting up the necessary configuration file in your home directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter your username: ")
		user, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading username: %v\n", err)
			return
		}
		user = strings.TrimSpace(user)
		if user == "" {
			fmt.Println("Username cannot be empty")
			return
		}

		fmt.Print("Enter the folder to save logs: ")
		logDir, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading log directory: %v\n", err)
			return
		}
		logDir = strings.TrimSpace(logDir)
		if logDir == "" {
			fmt.Println("Log directory cannot be empty")
			return
		}

		fmt.Print("Enter the branch to monitor (e.g., main): ")
		branch, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading branch: %v\n", err)
			return
		}
		branch = strings.TrimSpace(branch)
		if branch == "" {
			fmt.Println("Branch cannot be empty")
			return
		}

		fmt.Print("Enable Web GUI? (y/n): ")
		enableGUIStr, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading GUI preference: %v\n", err)
			return
		}
		enableGUI := strings.TrimSpace(strings.ToLower(enableGUIStr)) == "y"

		var port int
		if enableGUI {
			fmt.Print("Enter the port for the Web GUI (e.g., 9099): ")
			portStr, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading port: %v\n", err)
				return
			}
			portStr = strings.TrimSpace(portStr)
			if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
				fmt.Printf("Invalid port number: %v\n", err)
				return
			}
			if port < 1024 || port > 65535 {
				fmt.Println("Port must be between 1024 and 65535")
				return
			}
		}

		cfg := config.GlobalConfig{
			User:   user,
			LogDir: logDir,
			Branch: branch,
			Plugins: map[string]bool{
				"shell":  true,
				"docker": true,
			},
			GUI: config.GUIConfig{
				Enabled: enableGUI,
				Port:    port,
			},
		}

		if err := config.SaveGlobalConfig(cfg); err != nil {
			fmt.Println("Error saving configuration:", err)
			return
		}

		fmt.Println("Configuration saved successfully!")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}