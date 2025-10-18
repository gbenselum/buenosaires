// Package main is the entry point for the Buenos Aires application.
// Buenos Aires is a GitOps tool for monitoring Git repositories and
// automatically executing shell scripts that are committed to a monitored branch.
package main

import (
	"buenosaires/cmd"
)

// main is the application entry point that delegates execution to the cmd package.
func main() {
	cmd.Execute()
}