package proposal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/models"
	"github.com/stretchr/testify/assert"
)

// MockApprover はテスト用の承認フローモック
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

// MockClient はテスト用のAPIクライアントモック
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

func TestManager_GenerateProposal(t *testing.T) {
	// テストファイルの準備
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Hello World"), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		filePath  string
		mode      ApplyMode
		response  string
		apiErr    error
		approved  bool
		wantError bool
	}{
		{
			name:      "正常系：パッチモード",
			filePath:  testFile,
			mode:      ApplyModePatch,
			response:  "Hello Go",
			apiErr:    nil,
			approved:  true,
			wantError: false,
		},
		{
			name:      "正常系：全体書き換えモード",
			filePath:  testFile,
			mode:      ApplyModeFull,
			response:  "Hello Go",
			apiErr:    nil,
			approved:  true,
			wantError: false,
		},
		{
			name:      "異常系：ファイルが存在しない",
			filePath:  "nonexistent.txt",
			mode:      ApplyModePatch,
			response:  "",
			apiErr:    nil,
			approved:  true,
			wantError: true,
		},
		{
			name:      "異常系：APIエラー",
			filePath:  testFile,
			mode:      ApplyModePatch,
			response:  "",
			apiErr:    errors.New("API error"),
			approved:  true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				response: tt.response,
				err:      tt.apiErr,
			}
			manager := NewManager(mockClient, &MockApprover{approved: tt.approved})
			proposal, err := manager.GenerateProposal(tt.filePath, tt.mode)

			if tt.wantError {
				assert.Error(t, err)
				if !os.IsNotExist(err) { // ファイルが存在しないエラー以外の場合
					assert.Nil(t, proposal)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, proposal)
				assert.Equal(t, tt.filePath, proposal.FilePath)
				assert.Equal(t, tt.mode, proposal.ApplyMode)
			}
		})
	}
}

func TestManager_ApplyProposal(t *testing.T) {
	// テストファイルの準備
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Hello World"), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		proposal   *CodeProposal
		approved   bool
		approveErr error
		wantError  bool
	}{
		{
			name: "正常系：承認された変更",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     testFile,
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
			},
			approved:   true,
			approveErr: nil,
			wantError:  false,
		},
		{
			name: "異常系：承認されなかった変更",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     testFile,
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
			},
			approved:   false,
			approveErr: nil,
			wantError:  true,
		},
		{
			name: "異常系：承認プロセスでエラー",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     testFile,
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
			},
			approved:   false,
			approveErr: errors.New("approval error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				response: "",
				err:      nil,
			}
			approver := &MockApprover{
				approved: tt.approved,
				err:      tt.approveErr,
			}

			manager := NewManager(mockClient, approver)
			err := manager.ApplyProposal(tt.proposal)

			if tt.wantError {
				assert.Error(t, err)
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
