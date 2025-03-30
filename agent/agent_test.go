package agent

import (
	"context"
	"testing"
	"time"
)

// MockLLMClient はLLMClientのモック実装です
type MockLLMClient struct {
	responses []string
	current   int
}

func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if m.current >= len(m.responses) {
		return "", nil
	}
	response := m.responses[m.current]
	m.current++
	return response, nil
}

func (m *MockLLMClient) Close() {}

func TestAgent_SendMessage(t *testing.T) {
	// テストケースの準備
	tests := []struct {
		name     string
		messages []string
		want     string
	}{
		{
			name: "基本的な応答",
			messages: []string{
				"ファイル一覧を取得しました",
			},
			want: "ファイル一覧を取得しました",
		},
		{
			name: "ファイル読み取り",
			messages: []string{
				"ファイルの内容を読み取りました",
			},
			want: "ファイルの内容を読み取りました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モッククライアントの作成
			mock := &MockLLMClient{
				responses: tt.messages,
			}

			// エージェントの作成
			a, err := NewAgent(context.Background(), mock)
			if err != nil {
				t.Fatalf("NewAgent() error = %v", err)
			}

			// テスト実行
			ctx := context.Background()
			response, err := a.SendMessage(ctx, "テストメッセージ")

			// 結果の検証
			if err != nil {
				t.Errorf("SendMessage() error = %v", err)
				return
			}

			if response != tt.want {
				t.Errorf("SendMessage() = %v, want %v", response, tt.want)
			}
		})
	}
}

func TestAgent_buildPrompt(t *testing.T) {
	systemPrompt := `あなたはコードを読み取って解説を行うAIエージェントです。
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
</complete>`

	// テストケースの準備
	tests := []struct {
		name      string
		messages  []Message
		toolCalls []ToolCall
		want      string
	}{
		{
			name:      "空の会話",
			messages:  []Message{},
			toolCalls: []ToolCall{},
			want:      systemPrompt + "\n\n",
		},
		{
			name: "会話履歴あり",
			messages: []Message{
				{
					Role:      "user",
					Content:   "こんにちは",
					Timestamp: time.Now(),
				},
				{
					Role:      "assistant",
					Content:   "こんにちは！",
					Timestamp: time.Now(),
				},
			},
			toolCalls: []ToolCall{},
			want:      systemPrompt + "\n\nuser: こんにちは\nassistant: こんにちは！\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// エージェントの作成
			mock := &MockLLMClient{}
			a, err := NewAgent(context.Background(), mock)
			if err != nil {
				t.Fatalf("NewAgent() error = %v", err)
			}

			// 会話履歴とツール呼び出し履歴を設定
			a.conversation = tt.messages
			a.toolCalls = tt.toolCalls

			// テスト実行
			got := a.buildPrompt()

			// 結果の検証
			if got != tt.want {
				t.Errorf("buildPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
