package agent

import (
	"context"

	"github.com/pankona/makasero/tools"
)

// GeminiClientInterface はGeminiクライアントのインターフェースを定義します
type GeminiClientInterface interface {
	ProcessMessage(ctx context.Context, message string) (string, error)
	RegisterTool(tool tools.Tool)
	Close()
}

// Agent はAIエージェントのコアロジックを実装します
type Agent struct {
	gemini GeminiClientInterface
}

// New は新しいAgentを作成します
func New(apiKey string) (*Agent, error) {
	gemini, err := NewGeminiClient(apiKey)
	if err != nil {
		return nil, err
	}

	return &Agent{
		gemini: gemini,
	}, nil
}

// RegisterTool はツールを登録します
func (a *Agent) RegisterTool(tool tools.Tool) {
	a.gemini.RegisterTool(tool)
}

// Process はユーザーの入力を処理し、適切なアクションを実行します
func (a *Agent) Process(input string) (string, error) {
	ctx := context.Background()
	return a.gemini.ProcessMessage(ctx, input)
}

// Close はAgentを閉じます
func (a *Agent) Close() {
	a.gemini.Close()
}
