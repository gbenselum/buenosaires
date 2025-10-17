// Package cmd provides the command-line interface for Buenos Aires.
// It implements the core commands using the Cobra library, including:
//   - install: Setup and configuration
//   - run: Start the repository monitor
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command that serves as the entry point for all subcommands.
var rootCmd = &cobra.Command{
	Use:   "buenosaires",
	Short: "A tool to monitor a repository",
	Long:  `buenosaires is a Go-based tool for monitoring repositories and running plugins.`,
}

// Execute runs the root command and handles any errors that occur during execution.
// This is called by main.main() and only needs to be called once by the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}