package chat

import (
	"errors"
	"os"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/chat/detector"
	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/chat/handler"
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
	tempDir, err := os.MkdirTemp("", "roo-test-*")
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// カスタムExecutorの作成
			executor := &Executor{
				client:   tt.client,
				detector: detector.NewProposalDetector(),
				handler:  newTestHandler(tt.approved, nil, nil),
			}

			response, err := executor.Execute(tt.input, tt.targetFile)
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
