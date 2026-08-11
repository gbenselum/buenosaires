package shell

import (
	"os"
	"strings"
	"testing"
)

// writeTempScript creates a temporary shell script and registers its cleanup.
func writeTempScript(t *testing.T, name, content string) string {
	t.Helper()

	f, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return f.Name()
}

func TestShellPlugin_LintAndValidate(t *testing.T) {
	plugin := ShellPlugin{}

	t.Run("valid script", func(t *testing.T) {
		path := writeTempScript(t, "valid-*.sh", "#!/bin/bash\necho 'hello'\n")
		output, err := plugin.LintAndValidate(path)
		if err != nil {
			t.Errorf("Expected no error for valid script, but got: %v", err)
		}
		if !strings.Contains(output, "Syntax check passed") {
			t.Errorf("Expected output to contain 'Syntax check passed', but got: %s", output)
		}
		if !strings.Contains(output, "Linting completed") {
			t.Errorf("Expected output to contain 'Linting completed', but got: %s", output)
		}
	})

	t.Run("invalid syntax", func(t *testing.T) {
		path := writeTempScript(t, "invalid-syntax-*.sh", "#!/bin/bash\necho 'hello' &&\n")
		if _, err := plugin.LintAndValidate(path); err == nil {
			t.Error("Expected an error for invalid syntax, but got none")
		}
	})

	t.Run("shellcheck warning is non-fatal", func(t *testing.T) {
		path := writeTempScript(t, "shellcheck-*.sh", "#!/bin/bash\ncd /tmp\nls\n")
		output, err := plugin.LintAndValidate(path)
		if err != nil {
			t.Errorf("Expected no error for shellcheck warning, but got: %v", err)
		}
		if !strings.Contains(output, "SC2164") {
			t.Errorf("Expected output to contain shellcheck warning 'SC2164', but got: %s", output)
		}
	})
}

func TestShellPlugin_Run(t *testing.T) {
	t.Run("run without sudo", func(t *testing.T) {
		path := writeTempScript(t, "test-script-*.sh", "#!/bin/bash\necho 'hello'\n")
		plugin := ShellPlugin{}
		output, err := plugin.Run(path, false)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if !strings.Contains(output, "hello") {
			t.Errorf("Expected output to contain 'hello', but got: %s", output)
		}
	})
}

func TestShellPlugin_AssetRoundTrip(t *testing.T) {
	t.Cleanup(func() { _ = os.RemoveAll("plugins") })

	plugin := ShellPlugin{}
	t.Run("nested script name", func(t *testing.T) {
		asset, err := plugin.LoadAsset("shell/deploy.sh")
		if err != nil {
			t.Fatalf("LoadAsset error: %v", err)
		}
		if asset.Generation != 0 {
			t.Fatalf("expected fresh asset generation 0, got %d", asset.Generation)
		}
		asset.Generation = 1
		asset.Status = "success"
		if err := plugin.SaveAsset("shell/deploy.sh", asset); err != nil {
			t.Fatalf("SaveAsset error: %v", err)
		}
		loaded, err := plugin.LoadAsset("shell/deploy.sh")
		if err != nil {
			t.Fatalf("reload error: %v", err)
		}
		if loaded.Generation != 1 || loaded.Status != "success" {
			t.Errorf("unexpected asset after reload: %+v", loaded)
		}
	})

	t.Run("rejects traversal", func(t *testing.T) {
		if _, err := plugin.LoadAsset("../evil.sh"); err == nil {
			t.Error("expected error for traversal script name, got none")
		}
	})
}
