// Package shell provides a plugin for validating and executing shell scripts.
// It performs syntax checking with bash -n and linting with shellcheck before execution.
package shell

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ShellPlugin implements the plugin interface for handling shell scripts.
// It provides methods for linting/validation and execution.
type ShellPlugin struct{}

// LintAndValidate performs validation and linting on a shell script.
// It performs two checks:
//  1. Syntax validation using bash -n (non-execution mode)
//  2. Linting with shellcheck to identify potential issues
//
// Returns the combined output from both tools and any errors encountered.
// Shellcheck warnings (exit code 1) are not treated as fatal errors.
func (p *ShellPlugin) LintAndValidate(scriptPath string) (string, error) {
	var finalOutput bytes.Buffer

	// Step 1: Syntax check using bash in no-execution mode
	cmdBash := exec.Command("bash", "-n", scriptPath)
	bashOutput, err := cmdBash.CombinedOutput()
	finalOutput.Write(bashOutput)
	if err != nil {
		return finalOutput.String(), fmt.Errorf("syntax check failed: %w", err)
	}
	finalOutput.WriteString("Syntax check passed.\n")

	// Step 2: Run shellcheck for static analysis and best practices
	cmdShellcheck := exec.Command("shellcheck", "-s", "bash", scriptPath)
	shellcheckOutput, err := cmdShellcheck.CombinedOutput()
	finalOutput.Write(shellcheckOutput)

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Shellcheck returns exit code 1 for warnings (non-fatal)
			// Only treat exit codes > 1 as fatal errors
			if exitError.ExitCode() > 1 {
				return finalOutput.String(), fmt.Errorf("shellcheck failed with exit code %d: %w", exitError.ExitCode(), err)
			}
		} else {
			// Command execution failed (e.g., shellcheck not installed)
			return finalOutput.String(), fmt.Errorf("failed to run shellcheck: %w", err)
		}
	}
	finalOutput.WriteString("Linting completed.\n")

	return finalOutput.String(), nil
}

// Run executes a shell script using bash.
// Parameters:
//   - scriptPath: Path to the shell script to execute
//   - allowSudo: If true, the script is executed with sudo privileges
//
// Returns the combined stdout and stderr output, and any execution error.
func (p *ShellPlugin) Run(scriptPath string, allowSudo bool) (string, error) {
	var cmd *exec.Cmd
	// Execute with or without sudo based on configuration
	if allowSudo {
		cmd = exec.Command("sudo", "bash", scriptPath)
	} else {
		cmd = exec.Command("bash", scriptPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}