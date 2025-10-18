package config

import (
	"os"
	"reflect"
	"testing"
)

func TestSaveAndLoadGlobalConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-config")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory to use the temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Define a sample config
	expectedConfig := GlobalConfig{
		User:   "testuser",
		LogDir: "/tmp/logs",
	}

	// Save the config
	err = SaveGlobalConfig(expectedConfig)
	if err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	// Verify the config file was created
	configPath, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("Failed to get config file path: %v", err)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Load the config
	loadedConfig, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to load global config: %v", err)
	}

	// Compare the loaded config with the original
	if !reflect.DeepEqual(expectedConfig, loadedConfig) {
		t.Errorf("Loaded config does not match expected config.\nExpected: %+v\nGot:      %+v", expectedConfig, loadedConfig)
	}
}