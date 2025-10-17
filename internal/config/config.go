// Package config provides configuration management for Buenos Aires.
// It handles both global configuration (stored in ~/.buenosaires/config.toml)
// and repository-specific configuration (stored in the repository's config.toml).
package config

import (
	"os"
	"path/filepath"

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
	User    string          `toml:"user"`    // Default user for running scripts
	LogDir  string          `toml:"log_dir"` // Default directory for storing logs
	Branch  string          `toml:"branch"`  // Git branch to monitor
	Plugins map[string]bool `toml:"plugins"` // Enabled plugins (e.g., "shell")
	GUI     GUIConfig       `toml:"gui"`     // Web GUI configuration
}

// RepoConfig holds configuration specific to a repository.
// This is stored in the repository's config.toml file and overrides global settings.
type RepoConfig struct {
	User      string `toml:"user"`       // User to run scripts as (overrides global)
	LogDir    string `toml:"log_dir"`    // Log directory (overrides global)
	AllowSudo bool   `toml:"allow_sudo"` // Whether scripts can use sudo
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
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	}

	// Create the config file
	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(config)
}

// LoadRepoConfig loads the repository-specific configuration from config.toml in the repo directory.
// Returns an error if the file cannot be read or parsed.
func LoadRepoConfig(repoPath string) (RepoConfig, error) {
	var config RepoConfig
	configFile := filepath.Join(repoPath, "config.toml")
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return config, err
	}
	return config, nil
}