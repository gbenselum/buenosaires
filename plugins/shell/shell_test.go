package shell

import (
	"os"
	"strings"
	"testing"
)

func TestShellPlugin_LintAndValidate(t *testing.T) {
	// Test case 1: Valid script
	validScript := "#!/bin/bash\necho 'hello'"
	tmpfileValid, err := os.CreateTemp("", "valid-*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfileValid.Name())
	if _, err := tmpfileValid.Write([]byte(validScript)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfileValid.Close()

	plugin := ShellPlugin{}
	output, err := plugin.LintAndValidate(tmpfileValid.Name())
	if err != nil {
		t.Errorf("Expected no error for valid script, but got: %v", err)
	}
	if !strings.Contains(output, "Syntax check passed") {
		t.Errorf("Expected output to contain 'Syntax check passed', but got: %s", output)
	}
	if !strings.Contains(output, "Linting completed") {
		t.Errorf("Expected output to contain 'Linting completed', but got: %s", output)
	}

	// Test case 2: Invalid syntax
	invalidSyntaxScript := "#!/bin/bash\necho 'hello' &&"
	tmpfileInvalidSyntax, err := os.CreateTemp("", "invalid-syntax-*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfileInvalidSyntax.Name())
	if _, err := tmpfileInvalidSyntax.Write([]byte(invalidSyntaxScript)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfileInvalidSyntax.Close()

	_, err = plugin.LintAndValidate(tmpfileInvalidSyntax.Name())
	if err == nil {
		t.Error("Expected an error for invalid syntax, but got none")
	}

	// Test case 3: Shellcheck warning (should not fail)
	shellcheckWarningScript := "#!/bin/bash\ncd /tmp\nls"
	tmpfileShellcheck, err := os.CreateTemp("", "shellcheck-*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfileShellcheck.Name())
	if _, err := tmpfileShellcheck.Write([]byte(shellcheckWarningScript)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfileShellcheck.Close()

	output, err = plugin.LintAndValidate(tmpfileShellcheck.Name())
	if err != nil {
		t.Errorf("Expected no error for shellcheck warning, but got: %v", err)
	}
	if !strings.Contains(output, "SC2164") {
		t.Errorf("Expected output to contain shellcheck warning 'SC2164', but got: %s", output)
	}
}

func TestShellPlugin_Run(t *testing.T) {
	// Test case 1: Running without sudo
	scriptContent := "#!/bin/bash\necho 'hello'"
	tmpfile, err := os.CreateTemp("", "test-script-*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(scriptContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	plugin := ShellPlugin{}
	output, err := plugin.Run(tmpfile.Name(), false)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if !strings.Contains(output, "hello") {
		t.Errorf("Expected output to contain 'hello', but got: %s", output)
	}
}