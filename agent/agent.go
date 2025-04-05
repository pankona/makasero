package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pankona/makasero/tools"
)

// Agent はAIエージェントのコアロジックを実装します
type Agent struct {
	gemini *GeminiClient
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

	// Geminiにプロンプトを送信
	response, err := a.gemini.GenerateContent(ctx, input)
	if err != nil {
		return "", err
	}

	// レスポンスがJSON形式かどうかを確認
	if strings.HasPrefix(response, "{") {
		var toolCall struct {
			Tool string                 `json:"tool"`
			Args map[string]interface{} `json:"args"`
		}

		if err := json.Unmarshal([]byte(response), &toolCall); err != nil {
			return "", fmt.Errorf("failed to parse tool call: %v", err)
		}

		// ツールを実行
		tool, ok := a.gemini.tools[toolCall.Tool]
		if !ok {
			return "", fmt.Errorf("unknown tool: %s", toolCall.Tool)
		}

		result, err := tool.Execute(toolCall.Args)
		if err != nil {
			return "", fmt.Errorf("tool execution failed: %v", err)
		}

		// ツールの実行結果をGeminiに送信
		followUpPrompt := fmt.Sprintf("Tool execution result:\n%s\n\nPlease provide a final answer to the user's original input: %s", result, input)

		return a.gemini.GenerateContent(ctx, followUpPrompt)
	}

	// 直接の回答を返す
	return response, nil
}

// Close はAgentを閉じます
func (a *Agent) Close() {
	a.gemini.Close()
}
