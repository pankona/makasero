package agent

import (
	"context"
	"fmt"
	"time"
)

// Message は会話の1メッセージを表します
type Message struct {
	Role      string    // "user" または "assistant"
	Content   string    // メッセージの内容
	Timestamp time.Time // タイムスタンプ
}

// ToolCall はツールの呼び出しを表します
type ToolCall struct {
	Tool       string                 // ツール名
	Parameters map[string]interface{} // パラメータ
	Result     string                 // 実行結果
}

// LLMClient はLLMとの通信を抽象化するインターフェースです
type LLMClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	Close()
}

// Agent はAIエージェントの基本構造を表します
type Agent struct {
	client       LLMClient
	conversation []Message
	toolCalls    []ToolCall
	systemPrompt string
}

// NewAgent は新しいエージェントを作成します
func NewAgent(ctx context.Context, client LLMClient) (*Agent, error) {
	return &Agent{
		client:       client,
		conversation: make([]Message, 0),
		toolCalls:    make([]ToolCall, 0),
		systemPrompt: `あなたはコードを読み取って解説を行うAIエージェントです。
以下のツールを使って、ユーザーの要求に応じてコードの解説を行ってください。

# ListFile
ディレクトリ内のファイル一覧を取得します。
<list_file>
<path>ディレクトリのパス</path>
<recursive>true または false</recursive>
</list_file>

# ReadFile
ファイルの内容を読み取ります。
<read_file>
<path>ファイルのパス</path>
</read_file>

# Complete
タスク完了を示します。
<complete>
<summary>タスクの結果サマリー</summary>
</complete>`,
	}, nil
}

// SendMessage はユーザーからのメッセージを処理し、応答を返します
func (a *Agent) SendMessage(ctx context.Context, content string) (string, error) {
	// ユーザーメッセージを追加
	a.conversation = append(a.conversation, Message{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	})

	// 会話履歴を構築
	prompt := a.buildPrompt()

	// LLMに送信
	response, err := a.client.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	// アシスタントの応答を会話履歴に追加
	a.conversation = append(a.conversation, Message{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
	})

	return response, nil
}

// buildPrompt は会話履歴からプロンプトを構築します
func (a *Agent) buildPrompt() string {
	prompt := a.systemPrompt + "\n\n"

	// 会話履歴を追加
	for _, msg := range a.conversation {
		prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	// ツール呼び出し履歴を追加
	if len(a.toolCalls) > 0 {
		prompt += "\nTool Calls:\n"
		for _, call := range a.toolCalls {
			prompt += fmt.Sprintf("- %s: %v\n", call.Tool, call.Parameters)
			prompt += fmt.Sprintf("  Result: %s\n", call.Result)
		}
	}

	return prompt
}

// Close はエージェントのリソースを解放します
func (a *Agent) Close() {
	if a.client != nil {
		a.client.Close()
	}
}
