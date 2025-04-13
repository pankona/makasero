package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

func DefaultConfig() *Config {
	return &Config{
		MCPServers: map[string]MCPServerConfig{
			"claude": {
				Command: "claude",
				Args:    []string{"mcp", "serve"},
				Env:     map[string]string{},
			},
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		path = filepath.Join(homeDir, ".makasero", "config.json")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	if len(config.MCPServers) == 0 {
		return DefaultConfig(), nil
	}

	return &config, nil
}

func SaveConfig(config *Config, path string) error {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %v", err)
		}
		configDir := filepath.Join(homeDir, ".makasero")
		
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}
		
		path = filepath.Join(configDir, "config.json")
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
