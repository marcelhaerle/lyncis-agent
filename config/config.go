package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func SaveConfig(path string, config *Config) error {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, configData, 0600)
}
