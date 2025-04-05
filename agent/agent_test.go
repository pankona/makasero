package agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/pankona/makasero/tools"
)

// MockGeminiClient はテスト用のモックGeminiクライアントです
type MockGeminiClient struct {
	response string
	err      error
}

func (m *MockGeminiClient) ProcessMessage(ctx context.Context, message string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *MockGeminiClient) RegisterTool(tool tools.Tool) {}

func (m *MockGeminiClient) Close() {}

// NewMockAgent はモックを使用するAgentを作成します
func NewMockAgent(response string, err error) *Agent {
	return &Agent{
		gemini: &MockGeminiClient{
			response: response,
			err:      err,
		},
	}
}

func TestAgent_Process(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		response string
		wantErr  bool
	}{
		{
			name:     "simple question",
			input:    "1+1は？",
			response: "2です",
			wantErr:  false,
		},
		{
			name:     "tool execution",
			input:    "現在のディレクトリの内容を表示してください",
			response: "ls コマンドの実行結果: ...",
			wantErr:  false,
		},
		{
			name:    "error case",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.wantErr {
				err = fmt.Errorf("test error")
			}
			agent := NewMockAgent(tt.response, err)

			response, err := agent.Process(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Agent.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && response != tt.response {
				t.Errorf("Agent.Process() = %v, want %v", response, tt.response)
			}
		})
	}
}

// MockTool はテスト用のモックツールです
type MockTool struct {
	name        string
	description string
	executed    bool
	response    string
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) Execute(args map[string]interface{}) (string, error) {
	m.executed = true
	return m.response, nil
}

func TestAgent_WithMockTool(t *testing.T) {
	mockTool := &MockTool{
		name:        "mockTool",
		description: "A mock tool for testing",
		response:    "Mock tool executed successfully",
	}

	agent := NewMockAgent("ツールの実行結果: "+mockTool.response, nil)
	agent.RegisterTool(mockTool)

	response, err := agent.Process("mockToolを使って何か実行してください")
	if err != nil {
		t.Errorf("Agent.Process() error = %v", err)
		return
	}

	if response == "" {
		t.Error("Agent.Process() returned empty response")
	}
}

func Example() {
	// Note: This is just an example and requires a valid API key to run
	apiKey := "YOUR_API_KEY"
	agent, err := New(apiKey)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}
	defer agent.Close()

	agent.RegisterTool(&tools.ExecCommand{})

	response, err := agent.Process("現在のディレクトリの内容を表示してください")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}
