package status

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	StatusPending   = "pending"
	StatusSuccess   = "success"
	StatusFailure   = "failure"
	StatusSkipped   = "skipped"
)

// ScriptStatus represents the status of a single script.
type ScriptStatus struct {
	LintStatus    string    `json:"lint_status"`
	TestStatus    string    `json:"test_status"`
	RunStatus     string    `json:"run_status"`
	Timestamp     time.Time `json:"timestamp"`
	OverallStatus string    `json:"overall_status"`
}

// Status represents the overall status of all scripts in the repository.
type Status struct {
	Scripts map[string]ScriptStatus `json:"scripts"`
}

// getStatusFilePath returns the path to the status.json file.
func getStatusFilePath(repoPath string) string {
	return filepath.Join(repoPath, ".buenosaires", "status.json")
}

// LoadStatus loads the status from the status.json file.
func LoadStatus(repoPath string) (*Status, error) {
	statusFilePath := getStatusFilePath(repoPath)
	if _, err := os.Stat(statusFilePath); os.IsNotExist(err) {
		return &Status{Scripts: make(map[string]ScriptStatus)}, nil
	}

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

// SaveStatus saves the status to the status.json file.
func (s *Status) SaveStatus(repoPath string) error {
	statusFilePath := getStatusFilePath(repoPath)
	buenosairesDir := filepath.Dir(statusFilePath)
	if _, err := os.Stat(buenosairesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(buenosairesDir, 0755); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statusFilePath, data, 0644)
}

// UpdateScriptStatus updates the status of a script.
func (s *Status) UpdateScriptStatus(scriptName, lintStatus, testStatus, runStatus, overallStatus string) {
	s.Scripts[scriptName] = ScriptStatus{
		LintStatus:    lintStatus,
		TestStatus:    testStatus,
		RunStatus:     runStatus,
		Timestamp:     time.Now(),
		OverallStatus: overallStatus,
	}
}