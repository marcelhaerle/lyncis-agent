package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	expectedConfig := &Config{
		AgentID: "test-agent-id",
		Token:   "test-secret-token",
	}

	// Test SaveConfig
	err := SaveConfig(configPath, expectedConfig)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Test LoadConfig
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedConfig.AgentID != expectedConfig.AgentID {
		t.Errorf("expected AgentID %s, got %s", expectedConfig.AgentID, loadedConfig.AgentID)
	}

	if loadedConfig.Token != expectedConfig.Token {
		t.Errorf("expected Token %s, got %s", expectedConfig.Token, loadedConfig.Token)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non_existent_config.json")

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected error when loading non-existent config file, got nil")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_config.json")

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("invalid json data"), 0600)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected error when loading invalid JSON config file, got nil")
	}
}
