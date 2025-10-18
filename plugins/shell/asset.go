// Package shell provides the shell plugin for Buenos Aires.
package shell

import "time"

// Asset holds the metadata for a shell script asset.
type Asset struct {
	Generation   int       `json:"generation"`
	LastRun      time.Time `json:"last_run"`
	LintPassed   bool      `json:"lint_passed"`
	TestsPassed  bool      `json:"tests_passed"`
	Event        string    `json:"event"`
	User         string    `json:"user"`
	RunDuration  Duration  `json:"run_duration"`
	Status       string    `json:"status"`
	CommitHash   string    `json:"commit_hash"`
}
