package main

import (
	"errors"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/models"
	"github.com/stretchr/testify/assert"
)

// MockAPIClient はテスト用のモックAPIクライアント
type MockAPIClient struct {
	response string
	err      error
}

func (m *MockAPIClient) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestExecuteExplain(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		response string
		err      error
		want     string
		wantErr  bool
	}{
		{
			name:     "正常系：コードの説明が成功",
			code:     "fmt.Println('Hello')",
			response: "This code prints 'Hello' to the console",
			err:      nil,
			want:     "This code prints 'Hello' to the console",
			wantErr:  false,
		},
		{
			name:     "異常系：APIエラー",
			code:     "fmt.Println('Hello')",
			response: "",
			err:      errors.New("API error"),
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockAPIClient{
				response: tt.response,
				err:      tt.err,
			}

			got, err := executeExplain(client, tt.code)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExecuteChat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		response string
		err      error
		want     string
		wantErr  bool
	}{
		{
			name:     "正常系：チャットメッセージが成功",
			input:    `[{"role":"user","content":"Hello"}]`,
			response: "Hi there!",
			err:      nil,
			want:     "Hi there!",
			wantErr:  false,
		},
		{
			name:     "異常系：不正なJSON",
			input:    `invalid json`,
			response: "",
			err:      nil,
			want:     "",
			wantErr:  true,
		},
		{
			name:     "異常系：APIエラー",
			input:    `[{"role":"user","content":"Hello"}]`,
			response: "",
			err:      errors.New("API error"),
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockAPIClient{
				response: tt.response,
				err:      tt.err,
			}

			got, err := executeChat(client, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		input    string
		response string
		err      error
		want     interface{}
		wantErr  bool
	}{
		{
			name:     "正常系：explainコマンド",
			command:  "explain",
			input:    "fmt.Println('Hello')",
			response: "This code prints 'Hello'",
			err:      nil,
			want:     "This code prints 'Hello'",
			wantErr:  false,
		},
		{
			name:     "正常系：chatコマンド",
			command:  "chat",
			input:    `[{"role":"user","content":"Hello"}]`,
			response: "Hi there!",
			err:      nil,
			want:     "Hi there!",
			wantErr:  false,
		},
		{
			name:     "異常系：不明なコマンド",
			command:  "unknown",
			input:    "",
			response: "",
			err:      nil,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockAPIClient{
				response: tt.response,
				err:      tt.err,
			}

			got, err := executeCommand(client, tt.command, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
