package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/cmd/roo-helper/models"
)

// MockClient は APIClient のモック実装
type MockClient struct {
	response string
	err      error
}

func (m *MockClient) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		input    string
		response string
		wantErr  bool
	}{
		{
			name:     "explain command",
			command:  "explain",
			input:    "test code",
			response: "explanation of test code",
			wantErr:  false,
		},
		{
			name:     "chat command",
			command:  "chat",
			input:    `[{"role":"user","content":"Hello"}]`,
			response: "Hello! How can I help you?",
			wantErr:  false,
		},
		{
			name:    "unknown command",
			command: "unknown",
			input:   "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				response: tt.response,
			}

			result, err := executeCommand(mockClient, tt.command, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result != tt.response {
					t.Errorf("executeCommand() = %v, want %v", result, tt.response)
				}
			}
		})
	}
}

func TestExecuteExplain(t *testing.T) {
	mockClient := &MockClient{
		response: "This is a test explanation",
	}

	result, err := executeExplain(mockClient, "test code")
	if err != nil {
		t.Errorf("executeExplain() error = %v", err)
		return
	}

	if result != "This is a test explanation" {
		t.Errorf("executeExplain() = %v, want %v", result, "This is a test explanation")
	}
}

func TestExecuteChat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid chat messages",
			input:   `[{"role":"user","content":"Hello"}]`,
			want:    "Hello! How can I help you?",
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   "invalid json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				response: tt.want,
			}

			result, err := executeChat(mockClient, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.want {
				t.Errorf("executeChat() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestOutputResponse(t *testing.T) {
	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	response := models.Response{
		Success: true,
		Data:    "test data",
	}

	outputResponse(response)

	// 標準出力を元に戻す
	w.Close()
	os.Stdout = oldStdout

	// 出力を読み取る
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// JSONをデコード
	var result models.Response
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
		return
	}

	if result.Success != response.Success {
		t.Errorf("outputResponse() success = %v, want %v", result.Success, response.Success)
	}

	if result.Data != response.Data {
		t.Errorf("outputResponse() data = %v, want %v", result.Data, response.Data)
	}
}