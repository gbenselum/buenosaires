package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "buenosaires",
	Short: "A tool to monitor a repository",
	Long:  `buenosaires is a Go-based tool for monitoring repositories and running plugins.`,
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}