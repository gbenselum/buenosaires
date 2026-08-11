package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

func TestIsShellScriptInFolder(t *testing.T) {
	tests := []struct {
		name       string
		scriptName string
		folder     string
		want       bool
	}{
		{"script in scanned folder", "shell/deploy.sh", "shell", true},
		{"nested script in scanned folder", "shell/sub/deploy.sh", "shell", true},
		{"outside scanned folder", "scripts/deploy.sh", "shell", false},
		{"folder prefix collision", "shellish/deploy.sh", "shell", false},
		{"not a shell script", "shell/readme.txt", "shell", false},
		{"top level script", "deploy.sh", "shell", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isShellScriptInFolder(tt.scriptName, tt.folder); got != tt.want {
				t.Errorf("isShellScriptInFolder(%q, %q) = %v, want %v", tt.scriptName, tt.folder, got, tt.want)
			}
		})
	}
}

func TestShouldProcessScript(t *testing.T) {
	tests := []struct {
		name       string
		action     merkletrie.Action
		scriptName string
		folder     string
		prevStatus string
		want       bool
	}{
		// Insertions are processed unless the script already succeeded.
		{"insert new script", merkletrie.Insert, "shell/deploy.sh", "shell", "", true},
		{"insert previously failed script", merkletrie.Insert, "shell/deploy.sh", "shell", "failure", true},
		{"insert previously successful script is skipped", merkletrie.Insert, "shell/deploy.sh", "shell", "success", false},
		{"insert successful script during initial sync is skipped", merkletrie.Insert, "shell/deploy.sh", "shell", "success", false},
		// Modifications always re-deploy (fixes and new versions).
		{"modify successful script", merkletrie.Modify, "shell/deploy.sh", "shell", "success", true},
		{"modify failed script (retry)", merkletrie.Modify, "shell/deploy.sh", "shell", "failure", true},
		{"modify unprocessed script", merkletrie.Modify, "shell/deploy.sh", "shell", "", true},
		// Deletes and non-script files are never processed.
		{"delete script", merkletrie.Delete, "shell/deploy.sh", "shell", "", false},
		{"insert outside folder", merkletrie.Insert, "scripts/deploy.sh", "shell", "", false},
		{"insert non-script in folder", merkletrie.Insert, "shell/readme.txt", "shell", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldProcessScript(tt.action, tt.scriptName, tt.folder, tt.prevStatus)
			if got != tt.want {
				t.Errorf("shouldProcessScript(%v, %q, %q, %q) = %v, want %v", tt.action, tt.scriptName, tt.folder, tt.prevStatus, got, tt.want)
			}
		})
	}
}

// createTestTree builds a temporary git repository containing the given files
// and returns its tree object.
func createTestTree(t *testing.T, files map[string]string) *object.Tree {
	t.Helper()

	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if parent := filepath.Dir(path); parent != dir {
			if err := os.MkdirAll(parent, 0o755); err != nil {
				t.Fatalf("failed to create dir for %s: %v", name, err)
			}
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	if _, err := w.Add("."); err != nil {
		t.Fatalf("failed to stage files: %v", err)
	}
	hash, err := w.Commit("test commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
	commit, err := repo.CommitObject(hash)
	if err != nil {
		t.Fatalf("failed to get commit: %v", err)
	}
	tree, err := commit.Tree()
	if err != nil {
		t.Fatalf("failed to get tree: %v", err)
	}
	return tree
}

func TestInitialSyncChanges(t *testing.T) {
	tree := createTestTree(t, map[string]string{
		"shell/deploy.sh":   "#!/bin/bash\necho hi\n",
		"shell/sub/util.sh": "#!/bin/bash\necho util\n",
		"shell/readme.md":   "docs\n",
		"scripts/other.sh":  "#!/bin/bash\necho other\n",
		"not-in-folder.txt": "x\n",
	})

	changes, err := initialSyncChanges(tree, "shell")
	if err != nil {
		t.Fatalf("initialSyncChanges error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("initialSyncChanges returned %d changes, want 2", len(changes))
	}

	got := map[string]bool{}
	for _, c := range changes {
		action, err := c.Action()
		if err != nil {
			t.Fatalf("change.Action() error: %v", err)
		}
		if action != merkletrie.Insert {
			t.Errorf("expected Insert action, got %v", action)
		}
		got[c.To.Name] = true
	}
	for _, want := range []string{"shell/deploy.sh", "shell/sub/util.sh"} {
		if !got[want] {
			t.Errorf("initialSyncChanges missing %q (got %v)", want, got)
		}
	}
	if got["scripts/other.sh"] || got["shell/readme.md"] || got["not-in-folder.txt"] {
		t.Errorf("initialSyncChanges included non-matching files: %v", got)
	}
}
