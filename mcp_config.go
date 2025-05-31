package makasero

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MCPConfig struct {
	SystemPrompt string                      `json:"systemPrompt,omitempty"`
	Purpose      string                      `json:"purpose,omitempty"`
	MCPServers   map[string]MCPServerConfig  `json:"mcpServers"`
}

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

func LoadMCPConfig(path string) (*MCPConfig, error) {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		path = filepath.Join(homeDir, ".makasero", "config.json")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create directory if it does not exist
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %v", err)
			}
		}

		// Define default configuration
		defaultConfig := []byte(`{
  "systemPrompt": "あなたはAIアシスタントです。\nユーザーからのタスクを実行し、タスクが完了したら必ず'complete'関数を呼び出してください。\n関数を呼び出す際は、関数名をテキストとして書くのではなく、実際に関数を呼び出してください。",
  "purpose": "ユーザーのタスクを効率的に実行するAIアシスタント", 
  "mcpServers": {
    "claude": {
      "command": "claude",
      "args": ["mcp", "serve"],
      "env": {}
    }
  }
}`)
		if err := os.WriteFile(path, defaultConfig, 0644); err != nil {
			return nil, fmt.Errorf("failed to write default config file: %v", err)
		}
		fmt.Printf("Default config file created at %s\n", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}
