package proposal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pankona/makasero/internal/models"
	"github.com/stretchr/testify/assert"
)

// MockAPIClient はテスト用のAPIクライアントモック
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

// MockApprover はテスト用の承認モック
type MockApprover struct {
	approved bool
	err      error
}

func (m *MockApprover) RequestApproval(proposal *CodeProposal) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.approved, nil
}

func TestParseAIResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		wantCode    string
		wantDesc    string
		wantErr     bool
		errContains string
	}{
		{
			name: "正常系：正しい形式のレスポンス",
			response: `---PROPOSAL---
コードを改善しました。
---CODE---
func main() {
	fmt.Println("Hello")
}
---END---`,
			wantCode: `func main() {
	fmt.Println("Hello")
}`,
			wantDesc: "コードを改善しました。",
			wantErr:  false,
		},
		{
			name:        "異常系：不正な形式",
			response:    "不正な形式のレスポンス",
			wantErr:     true,
			errContains: "不正なレスポンス形式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, desc, err := parseAIResponse(tt.response)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCode, code)
				assert.Equal(t, tt.wantDesc, desc)
			}
		})
	}
}

func TestBackupFunctionality(t *testing.T) {
	// テスト用のディレクトリとファイルの準備
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	testFile := filepath.Join(tmpDir, "test.go")
	originalContent := "package main\n\nfunc main() {}\n"

	// テストファイルの作成
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	// マネージャーの作成
	manager := NewManager(
		&MockAPIClient{},
		&MockApprover{approved: true},
		backupDir,
	)

	// バックアップの作成テスト
	backupPath, err := manager.createBackup(testFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, backupPath)

	// バックアップファイルの存在確認
	_, err = os.Stat(backupPath)
	assert.NoError(t, err)

	// バックアップ内容の確認
	content, err := os.ReadFile(backupPath)
	assert.NoError(t, err)
	assert.Equal(t, originalContent, string(content))

	// ファイルの変更
	newContent := "package main\n\nfunc main() { println('hello') }\n"
	err = os.WriteFile(testFile, []byte(newContent), 0644)
	assert.NoError(t, err)

	// バックアップからの復元テスト
	err = manager.restoreBackup(backupPath, testFile)
	assert.NoError(t, err)

	// 復元後の内容確認
	restoredContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, originalContent, string(restoredContent))
}

func TestApplyProposal(t *testing.T) {
	// テスト用のディレクトリとファイルの準備
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	testFile := filepath.Join(tmpDir, "test.go")
	originalContent := "package main\n\nfunc main() {}\n"

	// テストファイルの作成
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		proposal    *CodeProposal
		approved    bool
		wantErr     bool
		errContains string
	}{
		{
			name: "正常系：承認された変更",
			proposal: &CodeProposal{
				OriginalCode: originalContent,
				ProposedCode: "package main\n\nfunc main() { println('hello') }\n",
				FilePath:     testFile,
				ApplyMode:    ApplyModeFull,
			},
			approved: true,
			wantErr:  false,
		},
		{
			name: "異常系：承認されなかった変更",
			proposal: &CodeProposal{
				OriginalCode: originalContent,
				ProposedCode: "package main\n\nfunc main() { println('hello') }\n",
				FilePath:     testFile,
				ApplyMode:    ApplyModeFull,
			},
			approved:    false,
			wantErr:     true,
			errContains: "ユーザーが変更を承認しませんでした",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// マネージャーの作成
			manager := NewManager(
				&MockAPIClient{},
				&MockApprover{approved: tt.approved},
				backupDir,
			)

			err := manager.ApplyProposal(tt.proposal)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)

				// 変更が適用されたことを確認
				content, err := os.ReadFile(tt.proposal.FilePath)
				assert.NoError(t, err)
				assert.Equal(t, tt.proposal.ProposedCode, string(content))

				// バックアップが作成されたことを確認
				files, err := os.ReadDir(backupDir)
				assert.NoError(t, err)
				assert.NotEmpty(t, files)
			}
		})
	}
}

func TestGenerateProposal(t *testing.T) {
	// テスト用のディレクトリとファイルの準備
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "package main\n\nfunc main() {}\n"

	// テストファイルの作成
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		response   string
		apiErr     error
		mode       ApplyMode
		wantErr    bool
		errMessage string
	}{
		{
			name: "正常系：パッチモード",
			response: `---PROPOSAL---
コードを改善しました。
---CODE---
package main

func main() {
	fmt.Println("Hello")
}
---END---`,
			mode:    ApplyModePatch,
			wantErr: false,
		},
		{
			name: "正常系：全体書き換えモード",
			response: `---PROPOSAL---
コードを改善しました。
---CODE---
package main

func main() {
	fmt.Println("Hello")
}
---END---`,
			mode:    ApplyModeFull,
			wantErr: false,
		},
		{
			name:       "異常系：APIエラー",
			response:   "",
			apiErr:     errors.New("API error"),
			mode:       ApplyModePatch,
			wantErr:    true,
			errMessage: "AI提案の生成に失敗",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockAPIClient{
				response: tt.response,
				err:      tt.apiErr,
			}

			manager := NewManager(mockClient, &MockApprover{}, "") // バックアップディレクトリなし
			proposal, err := manager.GenerateProposal(testFile, tt.mode)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
				assert.Nil(t, proposal)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, proposal)
				assert.Equal(t, testFile, proposal.FilePath)
				assert.Equal(t, tt.mode, proposal.ApplyMode)
				assert.Equal(t, originalContent, proposal.OriginalCode)
				assert.NotEmpty(t, proposal.ProposedCode)
				assert.NotEmpty(t, proposal.Description)
			}
		})
	}
}

func TestApplyProposalWithoutBackup(t *testing.T) {
	tests := []struct {
		name       string
		proposal   *CodeProposal
		approved   bool
		wantErr    bool
		errMessage string
	}{
		{
			name: "正常系：承認された変更",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
			},
			approved: true,
			wantErr:  false,
		},
		{
			name: "異常系：承認されなかった変更",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
			},
			approved:   false,
			wantErr:    true,
			errMessage: "ユーザーが変更を承認しませんでした",
		},
		{
			name: "異常系：承認プロセスでエラー",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
			},
			approved:   false,
			wantErr:    true,
			errMessage: "承認プロセスでエラー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockAPIClient{
				response: "",
				err:      nil,
			}

			var approver *MockApprover
			if tt.name == "異常系：承認プロセスでエラー" {
				approver = &MockApprover{approved: tt.approved, err: errors.New("承認プロセスでエラー")}
			} else {
				approver = &MockApprover{approved: tt.approved}
			}

			manager := NewManager(mockClient, approver, "") // バックアップディレクトリなし
			err := manager.ApplyProposal(tt.proposal)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				// 変更が適用されたことを確認
				content, err := os.ReadFile(tt.proposal.FilePath)
				assert.NoError(t, err)
				if tt.proposal.ApplyMode == ApplyModePatch {
					// パッチモードの場合は差分が正しく適用されていることを確認
					assert.Equal(t, tt.proposal.ProposedCode, string(content))
				}
			}
		})
	}
}
