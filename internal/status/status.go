// Package status manages the execution status of shell scripts.
// It tracks the lint, test, and run status for each script and persists
// this information in a .buenosaires/status.json file within the repository.
package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Status constants for tracking script execution state.
const (
	StatusPending = "pending" // Script is queued for processing
	StatusSuccess = "success" // Script processed successfully
	StatusFailure = "failure" // Script failed during processing
	StatusSkipped = "skipped" // Script processing was skipped
)

// ScriptStatus represents the execution status of a single shell script.
// It tracks the outcome of lint, test, and run phases, along with a timestamp.
type ScriptStatus struct {
	LintStatus    string    `json:"lint_status"`    // Result of linting (pending/success/failure)
	TestStatus    string    `json:"test_status"`    // Result of testing (pending/success/failure/skipped)
	RunStatus     string    `json:"run_status"`     // Result of execution (pending/success/failure)
	Timestamp     time.Time `json:"timestamp"`      // When the status was last updated
	OverallStatus string    `json:"overall_status"` // Overall result of all phases
}

// Status represents the overall status tracking for all scripts in the repository.
// It maps script names to their execution status.
type Status struct {
	Scripts map[string]ScriptStatus `json:"scripts"`
}

// getStatusFilePath returns the path to the status.json file within the repository.
// The file is stored in the .buenosaires directory.
func getStatusFilePath(repoPath string) (string, error) {
	// Sanitize the repo path to prevent directory traversal
	cleanRepoPath := filepath.Clean(repoPath)
	if cleanRepoPath != repoPath || repoPath == ".." || repoPath == "." {
		return "", fmt.Errorf("invalid repo path: %s", repoPath)
	}
	return filepath.Join(cleanRepoPath, ".buenosaires", "status.json"), nil
}

// LoadStatus loads the status from the status.json file in the repository.
// If the file doesn't exist, it returns a new empty Status object.
func LoadStatus(repoPath string) (*Status, error) {
	statusFilePath, err := getStatusFilePath(repoPath)
	if err != nil {
		return nil, err
	}
	// If the status file doesn't exist yet, return an empty status
	if _, err := os.Stat(statusFilePath); os.IsNotExist(err) {
		return &Status{Scripts: make(map[string]ScriptStatus)}, nil
	}

	// #nosec G304
	data, err := os.ReadFile(statusFilePath)
	if err != nil {
		return nil, err
	}

	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// SaveStatus persists the current status to the status.json file.
// It creates the .buenosaires directory if it doesn't exist.
func (s *Status) SaveStatus(repoPath string) error {
	statusFilePath, err := getStatusFilePath(repoPath)
	if err != nil {
		return err
	}
	buenosairesDir := filepath.Dir(statusFilePath)
	// Create the .buenosaires directory if it doesn't exist
	if _, err := os.Stat(buenosairesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(buenosairesDir, 0750); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statusFilePath, data, 0600)
}

// UpdateScriptStatus updates the status of a specific script.
// It creates a new ScriptStatus entry with the provided status values and current timestamp.
func (s *Status) UpdateScriptStatus(scriptName, lintStatus, testStatus, runStatus, overallStatus string) {
	s.Scripts[scriptName] = ScriptStatus{
		LintStatus:    lintStatus,
		TestStatus:    testStatus,
		RunStatus:     runStatus,
		Timestamp:     time.Now(),
		OverallStatus: overallStatus,
	}
}