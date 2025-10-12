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
		user, _ := reader.ReadString('\n')
		user = strings.TrimSpace(user)

		fmt.Print("Enter the folder to save logs: ")
		logDir, _ := reader.ReadString('\n')
		logDir = strings.TrimSpace(logDir)

		fmt.Print("Enter the branch to monitor (e.g., main): ")
		branch, _ := reader.ReadString('\n')
		branch = strings.TrimSpace(branch)

		cfg := config.GlobalConfig{
			User:   user,
			LogDir: logDir,
			Branch: branch,
			Plugins: map[string]bool{
				"shell": true,
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