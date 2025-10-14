package status

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
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
	LintStatus       string    `json:"lint_status"`
	TestStatus       string    `json:"test_status"`
	RunStatus        string    `json:"run_status"`
	Timestamp        time.Time `json:"timestamp"`
	OverallStatus    string    `json:"overall_status"`
	Generation       int       `json:"generation"`        // Increments each time the script is redeployed
	FirstDeployDate  time.Time `json:"first_deploy_date"` // Date when script was first deployed
	CurrentVersionDate time.Time `json:"current_version_date"` // Date of the current version
}

// Status represents the overall status of all scripts in the repository.
type Status struct {
	mu      sync.RWMutex
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
	if status.Scripts == nil {
		status.Scripts = make(map[string]ScriptStatus)
	}
	return &status, nil
}

// SaveStatus saves the status to the status.json file.
func (s *Status) SaveStatus(repoPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
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
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	existing, exists := s.Scripts[scriptName]
	
	var generation int
	var firstDeployDate time.Time
	
	if exists {
		// Script already exists, increment generation
		generation = existing.Generation + 1
		firstDeployDate = existing.FirstDeployDate
	} else {
		// New script
		generation = 1
		firstDeployDate = now
	}
	
	s.Scripts[scriptName] = ScriptStatus{
		LintStatus:         lintStatus,
		TestStatus:         testStatus,
		RunStatus:          runStatus,
		Timestamp:          now,
		OverallStatus:      overallStatus,
		Generation:         generation,
		FirstDeployDate:    firstDeployDate,
		CurrentVersionDate: now,
	}
}

// GetScriptStatus retrieves the status of a script in a thread-safe manner.
func (s *Status) GetScriptStatus(scriptName string) (ScriptStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, exists := s.Scripts[scriptName]
	return status, exists
}