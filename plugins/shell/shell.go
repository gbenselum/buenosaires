// Package shell provides a plugin for validating and executing shell scripts.
// It performs syntax checking with bash -n and linting with shellcheck before execution.
package shell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ShellPlugin implements the plugin interface for handling shell scripts.
// It provides methods for linting/validation and execution.
type ShellPlugin struct{}

// getAssetPath returns the path to the asset JSON file for a given script.
func (p *ShellPlugin) getAssetPath(scriptName string) (string, error) {
	assetsDir := "plugins/shell/assets"
	if err := os.MkdirAll(assetsDir, 0750); err != nil {
		return "", err
	}
	// Sanitize the script name to prevent directory traversal
	cleanScriptName := filepath.Clean(scriptName)
	if cleanScriptName != scriptName || scriptName == ".." || scriptName == "." {
		return "", fmt.Errorf("invalid script name: %s", scriptName)
	}
	return filepath.Join(assetsDir, cleanScriptName+".json"), nil
}

// LoadAsset loads the asset metadata for a given script.
func (p *ShellPlugin) LoadAsset(scriptName string) (Asset, error) {
	var asset Asset
	assetPath, err := p.getAssetPath(scriptName)
	if err != nil {
		return asset, err
	}

	// #nosec G304
	data, err := os.ReadFile(assetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Asset{Generation: 0}, nil // Not found is not an error, just means it's a new script
		}
		return asset, err
	}

	if err := json.Unmarshal(data, &asset); err != nil {
		return asset, err
	}
	return asset, nil
}

// SaveAsset saves the asset metadata for a given script.
func (p *ShellPlugin) SaveAsset(scriptName string, asset Asset) error {
	assetPath, err := p.getAssetPath(scriptName)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(asset, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(assetPath, data, 0600)
}

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

// UpdateAssetAfterRun updates the asset metadata after a script has been run.
func (p *ShellPlugin) UpdateAssetAfterRun(scriptName, user, commitHash, event string, lintPassed bool, runDuration time.Duration, runStatus string) error {
	asset, err := p.LoadAsset(scriptName)
	if err != nil {
		return err
	}

	asset.Generation++
	asset.LastRun = time.Now()
	asset.LintPassed = lintPassed
	// The shell plugin does not currently support running tests, so this is hardcoded to true.
	// In the future, this should be updated to reflect the actual test results.
	asset.TestsPassed = true
	asset.Event = event
	asset.User = user
	asset.RunDuration = Duration{runDuration}
	asset.Status = runStatus
	asset.CommitHash = commitHash

	return p.SaveAsset(scriptName, asset)
}