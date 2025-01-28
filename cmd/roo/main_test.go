package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/models"
)

// MockAPIClient はAPIClientのモック実装です
type MockAPIClient struct {
	response string
	err      error
}

func (m *MockAPIClient) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
	return m.response, m.err
}

func TestExecuteCommand(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "roo-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name       string
		command    string
		input      string
		targetFile string
		setup      func(t *testing.T, tempDir string) string // テスト前処理用関数
		response   string
		wantErr    bool
	}{
		{
			name:       "explain command",
			command:    "explain",
			input:      "func main() {}",
			targetFile: "",
			response:   "This is a simple main function.",
			wantErr:    false,
		},
		{
			name:       "chat command",
			command:    "chat",
			input:      "コードを改善してください",
			targetFile: "test.go",
			setup: func(t *testing.T, tempDir string) string {
				testFile := filepath.Join(tempDir, "test.go")
				if err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
					t.Fatalf("一時ファイルの作成に失敗: %v", err)
				}
				return testFile
			},
			response: "はい、改善案を提示します。",
			wantErr:  false,
		},
		{
			name:       "unknown command",
			command:    "unknown",
			input:      "",
			targetFile: "",
			response:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockAPIClient{
				response: tt.response,
				err:      nil,
			}

			// テストケース用のバックアップディレクトリを作成
			backupDir := filepath.Join(tempDir, tt.name)

			targetFile := tt.targetFile
			if tt.setup != nil {
				if tt.setup != nil {
					targetFile = tt.setup(t, tempDir)
				}
			}
			got, err := executeCommand(client, tt.command, tt.input, targetFile, backupDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.response {
				t.Errorf("executeCommand() = %v, want %v", got, tt.response)
			}
		})
	}
}

func TestExecuteExplain(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		response string
		wantErr  bool
	}{
		{
			name:     "simple code",
			code:     "func main() {}",
			response: "This is a simple main function.",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockAPIClient{
				response: tt.response,
				err:      nil,
			}

			got, err := executeExplain(client, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeExplain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.response {
				t.Errorf("executeExplain() = %v, want %v", got, tt.response)
			}
		})
	}
}

func TestExecuteChat(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "roo-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name       string
		input      string
		targetFile string
		setup      func(t *testing.T, tempDir string) string // テスト前処理用関数
		response   string
		wantErr    bool
	}{
		{
			name:       "simple chat",
			input:      "コードを改善してください",
			targetFile: "test.go",
			setup: func(t *testing.T, tempDir string) string {
				testFile := filepath.Join(tempDir, "test.go")
				if err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
					t.Fatalf("一時ファイルの作成に失敗: %v", err)
				}
				return testFile
			},
			response: "はい、改善案を提示します。",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockAPIClient{
				response: tt.response,
				err:      nil,
			}

			// テストケース用のバックアップディレクトリを作成
			backupDir := filepath.Join(tempDir, tt.name)

			targetFile := tt.targetFile
			if tt.setup != nil {
				targetFile = tt.setup(t, tempDir)
			}
			got, err := executeChat(client, tt.input, targetFile, backupDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.response {
				t.Errorf("executeChat() = %v, want %v", got, tt.response)
			}
		})
	}
}
