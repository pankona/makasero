package chat

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pankona/makasero/internal/chat/detector"
	"github.com/pankona/makasero/internal/chat/handler"
)

// モックチャットクライアント
type mockChatClient struct {
	response string
	err      error
}

func (m *mockChatClient) CreateChatCompletion(messages []Message) (string, error) {
	return m.response, m.err
}

// モックApprover
type mockApprover struct {
	approved bool
	err      error
}

func (m *mockApprover) GetApproval(*detector.Proposal) (bool, error) {
	return m.approved, m.err
}

// モックApplier
type mockApplier struct {
	err error
}

func (m *mockApplier) Apply(*detector.Proposal) error {
	return m.err
}

// テスト用のProposalHandlerを作成するヘルパー関数
func newTestHandler(approved bool, approverErr, applierErr error) *handler.ProposalHandler {
	return handler.NewProposalHandler(
		&mockApprover{approved: approved, err: approverErr},
		&mockApplier{err: applierErr},
	)
}

func TestExecutor_Execute(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "makasero-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name       string
		client     *mockChatClient
		input      string
		targetFile string
		approved   bool
		wantErr    bool
	}{
		{
			name: "正常系：提案なし",
			client: &mockChatClient{
				response: "はい、その実装で問題ありません。",
				err:      nil,
			},
			input:      "コードをレビューしてください",
			targetFile: "",
			approved:   false,
			wantErr:    false,
		},
		{
			name: "正常系：提案あり（承認）",
			client: &mockChatClient{
				response: `---PROPOSAL---
エラーハンドリングを改善します。
---FILE---
test.go
---DIFF---
テスト差分
---END---`,
				err: nil,
			},
			input:      "コードを改善してください",
			targetFile: "test.go",
			approved:   true,
			wantErr:    false,
		},
		{
			name: "正常系：提案あり（非承認）",
			client: &mockChatClient{
				response: `---PROPOSAL---
エラーハンドリングを改善します。
---FILE---
test.go
---DIFF---
テスト差分
---END---`,
				err: nil,
			},
			input:      "コードを改善してください",
			targetFile: "test.go",
			approved:   false,
			wantErr:    false,
		},
		{
			name: "異常系：APIエラー",
			client: &mockChatClient{
				response: "",
				err:      errors.New("API error"),
			},
			input:      "テスト",
			targetFile: "",
			approved:   false,
			wantErr:    true,
		},
	}

	// テストファイルの作成
	testFilePath := filepath.Join(tempDir, "test.go")
	testContent := `package main

func main() {}
`
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// カスタムExecutorの作成
			executor := &Executor{
				client:   tt.client,
				detector: detector.NewProposalDetector(),
				handler:  newTestHandler(tt.approved, nil, nil),
			}

			// テストケースのtargetFileを一時ディレクトリ内のファイルに更新
			targetFile := tt.targetFile
			if targetFile == "test.go" {
				targetFile = testFilePath
			}

			response, err := executor.Execute(tt.input, targetFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && response != tt.client.response {
				t.Errorf("Execute() response = %v, want %v", response, tt.client.response)
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name      string
		backupDir string
		wantErr   bool
	}{
		{
			name:      "正常系：バックアップディレクトリ指定あり",
			backupDir: os.TempDir(),
			wantErr:   false,
		},
		{
			name:      "正常系：バックアップディレクトリ指定なし",
			backupDir: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockChatClient{}
			executor, err := NewExecutor(client, tt.backupDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExecutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && executor == nil {
				t.Error("NewExecutor() returned nil executor")
			}
		})
	}
}

func TestCreatePrompt(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "makasero-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テストファイルの作成
	testFilePath := filepath.Join(tempDir, "test.go")
	testContent := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Print("Enter your message: ")
	var input string
	_, err := fmt.Scanf("%s\n", &input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Your message: %s\n", input)
}
`
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	tests := []struct {
		name        string
		input       string
		filePath    string
		wantErr     bool
		wantContent string
	}{
		{
			name:        "ファイルパスなし",
			input:       "コードを改善してください",
			filePath:    "",
			wantErr:     false,
			wantContent: "コードを改善してください",
		},
		{
			name:     "存在しないファイル",
			input:    "コードを改善してください",
			filePath: filepath.Join(tempDir, "nonexistent.go"),
			wantErr:  true,
		},
		{
			name:     "正常系：ファイルあり（エラーハンドリングの改善要求）",
			input:    "このコードを改善してください。特にエラーハンドリングの部分を見直してください。",
			filePath: testFilePath,
			wantErr:  false,
			wantContent: fmt.Sprintf("このコードを改善してください。特にエラーハンドリングの部分を見直してください。\n\n対象ファイル: %s\n\nコード：\n```go\n%s\n```",
				testFilePath, testContent),
		},
		{
			name:     "正常系：ファイルあり（一般的な改善要求）",
			input:    "このコードをよりGo言語のベストプラクティスに沿うように改善してください。",
			filePath: testFilePath,
			wantErr:  false,
			wantContent: fmt.Sprintf("このコードをよりGo言語のベストプラクティスに沿うように改善してください。\n\n対象ファイル: %s\n\nコード：\n```go\n%s\n```",
				testFilePath, testContent),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createPrompt(tt.input, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantContent {
				t.Errorf("createPrompt() = %v, want %v", got, tt.wantContent)
			}
		})
	}
}
