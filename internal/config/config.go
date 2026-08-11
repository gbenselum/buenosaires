// Package config provides configuration management for Buenos Aires.
// It handles both global configuration (stored in ~/.buenosaires/config.toml)
// and repository-specific configuration (stored in the repository's config.toml).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// GUIConfig holds the configuration for the web GUI that displays script logs.
type GUIConfig struct {
	Enabled bool `toml:"enabled"` // Whether the web GUI is enabled
	Port    int  `toml:"port"`    // Port number for the web server
}

// GlobalConfig holds the global configuration stored in ~/.buenosaires/config.toml.
// This configuration applies as the default for all repositories.
type GlobalConfig struct {
	User          string    `toml:"user"`
	LogDir        string    `toml:"log_dir"`
	Branch        string    `toml:"branch"`
	SyncInterval  int       `toml:"sync_interval"`
	GUI           GUIConfig `toml:"gui"`
	RepositoryURL string    `toml:"repository_url"`
	// AllowSudo is the host-operator gate for sudo execution. A repository's
	// config.toml cannot enable sudo on its own; both this setting AND the
	// repository's allow_sudo must be true for scripts to run with elevated
	// privileges. This prevents a malicious repo from escalating privileges.
	AllowSudo bool `toml:"allow_sudo"`
}

// PluginConfig holds configuration specific to a plugin.
type PluginConfig struct {
	Enabled      bool   `toml:"enabled"`
	FolderToScan string `toml:"folder_to_scan"`
}

// RepoConfig holds configuration specific to a repository.
// This is stored in the repository's config.toml file and overrides global settings.
type RepoConfig struct {
	User      string                  `toml:"user"`       // User to run scripts as (overrides global)
	LogDir    string                  `toml:"log_dir"`    // Log directory (overrides global)
	AllowSudo bool                    `toml:"allow_sudo"` // Whether scripts can use sudo
	Plugins   map[string]PluginConfig `toml:"plugins"`    // Per-plugin configuration
}

// GetConfigDir returns the path to the .buenosaires configuration directory in the user's home.
// This directory stores the global configuration file.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".buenosaires"), nil
}

// GetConfigFilePath returns the full path to the global config.toml file.
// The file is located at ~/.buenosaires/config.toml.
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.toml"), nil
}

// LoadGlobalConfig loads and parses the global configuration from ~/.buenosaires/config.toml.
// Returns an error if the file cannot be read or parsed.
func LoadGlobalConfig() (GlobalConfig, error) {
	var config GlobalConfig
	configFile, err := GetConfigFilePath()
	if err != nil {
		return config, err
	}

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return config, err
	}
	return config, nil
}

// SaveGlobalConfig saves the global configuration to ~/.buenosaires/config.toml.
// It creates the configuration directory if it doesn't exist.
func SaveGlobalConfig(config GlobalConfig) error {
	configFile, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create the configuration directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0750); err != nil {
			return err
		}
	}

	// Create the config file
	// #nosec G304
	file, err := os.Create(configFile)
	if err != nil {
		return err
	}

	// Encode first, then close. Returning the close error ensures a failed
	// flush is surfaced instead of silently losing the configuration.
	if err := toml.NewEncoder(file).Encode(config); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// sanitizeRepoPath validates and normalizes a repository path, rejecting
// absolute paths and directory traversal attempts. It returns the cleaned
// relative path.
func sanitizeRepoPath(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("empty repo path")
	}
	clean := filepath.Clean(repoPath)
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("absolute repo path not allowed: %s", repoPath)
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid repo path (traversal): %s", repoPath)
	}
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		if part == ".." {
			return "", fmt.Errorf("invalid repo path (traversal): %s", repoPath)
		}
	}
	return clean, nil
}

// LoadRepoConfig loads the repository-specific configuration from config.toml in the repo directory.
// Returns an error if the file cannot be read or parsed.
func LoadRepoConfig(repoPath string) (RepoConfig, error) {
	var config RepoConfig
	cleanRepoPath, err := sanitizeRepoPath(repoPath)
	if err != nil {
		return config, err
	}
	configFile := filepath.Join(cleanRepoPath, "config.toml")
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return config, err
	}
	return config, nil
}
