package shell

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ShellPlugin represents the shell script plugin.
type ShellPlugin struct{}

// LintAndValidate performs a dry run and linting on the script.
func (p *ShellPlugin) LintAndValidate(scriptPath string) (string, error) {
	var finalOutput bytes.Buffer

	// 1. Syntax check with bash -n
	cmdBash := exec.Command("bash", "-n", scriptPath)
	bashOutput, err := cmdBash.CombinedOutput()
	finalOutput.Write(bashOutput)
	if err != nil {
		return finalOutput.String(), fmt.Errorf("syntax check failed: %w", err)
	}
	finalOutput.WriteString("Syntax check passed.\n")

	// 2. Lint with shellcheck
	cmdShellcheck := exec.Command("shellcheck", "-s", "bash", scriptPath)
	shellcheckOutput, err := cmdShellcheck.CombinedOutput()
	finalOutput.Write(shellcheckOutput)

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// shellcheck returns exit code 1 for warnings. We don't treat this as a fatal error.
			if exitError.ExitCode() > 1 {
				return finalOutput.String(), fmt.Errorf("shellcheck failed with exit code %d: %w", exitError.ExitCode(), err)
			}
		} else {
			// Not an ExitError, so something else went wrong (e.g., command not found).
			return finalOutput.String(), fmt.Errorf("failed to run shellcheck: %w", err)
		}
	}
	finalOutput.WriteString("Linting completed.\n")

	return finalOutput.String(), nil
}

// Run executes the shell script plugin.
func (p *ShellPlugin) Run(scriptPath string, allowSudo bool) (string, error) {
	var cmd *exec.Cmd
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