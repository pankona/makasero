package main

import (
	"context"

	"github.com/pankona/makasero"
)

// Agentを作成するインターフェース
type AgentCreator interface {
	NewAgent(ctx context.Context, apiKey string, config *makasero.MCPConfig, opts ...makasero.AgentOption) (AgentProcessor, error)
}

// Agentの主要な処理を行うインターフェース
type AgentProcessor interface {
	ProcessMessage(ctx context.Context, userInput string) error
	Close() error
}

// セッションをロードするインターフェース
type SessionLoader interface {
	LoadSession(id string) (*makasero.Session, error)
}

// 設定をロードするインターフェース
type ConfigLoader interface {
	LoadMCPConfig(path string) (*makasero.MCPConfig, error)
}

// --- デフォルト実装 ---

// makasero パッケージの関数をラップするデフォルト実装
type defaultAgentCreator struct{}

func (d *defaultAgentCreator) NewAgent(ctx context.Context, apiKey string, config *makasero.MCPConfig, opts ...makasero.AgentOption) (AgentProcessor, error) {
	// makasero.NewAgent は *makasero.Agent を返す。これが AgentProcessor を満たすと仮定。
	agent, err := makasero.NewAgent(ctx, apiKey, config, opts...)
	if err != nil {
		return nil, err
	}
	return agent, nil // *makasero.Agent を AgentProcessor として返す
}

type defaultSessionLoader struct{}

func (d *defaultSessionLoader) LoadSession(id string) (*makasero.Session, error) {
	return makasero.LoadSession(id)
}

type defaultConfigLoader struct{}

func (d *defaultConfigLoader) LoadMCPConfig(path string) (*makasero.MCPConfig, error) {
	return makasero.LoadMCPConfig(path)
}
