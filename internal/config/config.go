package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// GUIConfig holds the configuration for the web GUI.
type GUIConfig struct {
	Enabled bool `toml:"enabled"`
	Port    int  `toml:"port"`
}

// GlobalConfig holds the configuration that is stored in the user's home directory.
type GlobalConfig struct {
	User    string         `toml:"user"`
	LogDir  string         `toml:"log_dir"`
	Branch  string         `toml:"branch"`
	Plugins map[string]bool `toml:"plugins"`
	GUI     GUIConfig      `toml:"gui"`
}

// DockerConfig holds Docker-specific configuration.
type DockerConfig struct {
	Enabled      bool   `toml:"enabled"`       // Enable Docker plugin
	AutoRun      bool   `toml:"auto_run"`      // Automatically run containers after building
	DefaultTag   string `toml:"default_tag"`   // Default tag for images (e.g., "latest")
	ImagePrefix  string `toml:"image_prefix"`  // Prefix for image names (e.g., "mycompany/")
}

// RepoConfig holds the configuration specific to a repository.
type RepoConfig struct {
	User      string       `toml:"user"`
	LogDir    string       `toml:"log_dir"`
	AllowSudo bool         `toml:"allow_sudo"`
	Docker    DockerConfig `toml:"docker"`
}

// GetConfigDir returns the path to the .buenosaires configuration directory.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".buenosaires"), nil
}

// GetConfigFilePath returns the path to the global config.toml file.
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.toml"), nil
}

// LoadGlobalConfig loads the global configuration from ~/.buenosaires/config.toml.
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
func SaveGlobalConfig(config GlobalConfig) error {
	configFile, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	}

	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(config)
}

// LoadRepoConfig loads the repository-specific configuration from the given path.
func LoadRepoConfig(repoPath string) (RepoConfig, error) {
	var config RepoConfig
	configFile := filepath.Join(repoPath, "config.toml")
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return config, err
	}
	return config, nil
}