package main

import (
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

func TestParseAIResponse(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		wantErr      bool
		wantProposal string
		wantCode     string
	}{
		{
			name: "正常系：正しい形式のレスポンス",
			response: `---PROPOSAL---
コードを改善しました。
エラーハンドリングを追加しました。
---CODE---
package main

func main() {
    if err := process(); err != nil {
        log.Fatal(err)
    }
}
---END---`,
			wantErr:      false,
			wantProposal: "コードを改善しました。\nエラーハンドリングを追加しました。",
			wantCode: `package main

func main() {
    if err := process(); err != nil {
        log.Fatal(err)
    }
}`,
		},
		{
			name:         "異常系：不正な形式",
			response:     "不正な形式のレスポンス",
			wantErr:      true,
			wantProposal: "",
			wantCode:     "",
		},
		{
			name: "異常系：セクションの欠落",
			response: `---PROPOSAL---
コードを改善しました。
---END---`,
			wantErr:      true,
			wantProposal: "",
			wantCode:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proposal, code, err := parseAIResponse(tt.response)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantProposal, proposal)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestCreateBackup(t *testing.T) {
	// テスト用のディレクトリとファイルを準備
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test content"

	err := os.WriteFile(testFile, []byte(content), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		backupDir string
		wantErr   bool
	}{
		{
			name:      "正常系：バックアップ作成",
			backupDir: backupDir,
			wantErr:   false,
		},
		{
			name:      "正常系：バックアップディレクトリなし",
			backupDir: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createBackup(testFile, tt.backupDir)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.backupDir != "" {
				// バックアップファイルの存在確認
				files, err := os.ReadDir(tt.backupDir)
				assert.NoError(t, err)
				assert.NotEmpty(t, files)

				// バックアップ内容の確認
				backupPath := filepath.Join(tt.backupDir, files[0].Name())
				backupContent, err := os.ReadFile(backupPath)
				assert.NoError(t, err)
				assert.Equal(t, content, string(backupContent))
			}
		})
	}
}

func TestApplyChanges(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "original content"
	newContent := "new content"

	// テストファイルの作成
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		filePath   string
		newContent string
		wantErr    bool
	}{
		{
			name:       "正常系：変更適用",
			filePath:   testFile,
			newContent: newContent,
			wantErr:    false,
		},
		{
			name:       "異常系：存在しないパス",
			filePath:   filepath.Join(tmpDir, "nonexistent", "test.txt"),
			newContent: newContent,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyChanges(tt.filePath, tt.newContent)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			content, err := os.ReadFile(tt.filePath)
			assert.NoError(t, err)
			assert.Equal(t, tt.newContent, string(content))
		})
	}
}

func TestExecuteChat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	backupDir := filepath.Join(tmpDir, "backups")
	originalContent := "package main\n\nfunc main() {}\n"

	// テストファイルの作成
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	assert.NoError(t, err)

	// chatApplyフラグの一時的な値を設定
	chatApplyValue := false
	chatApply = &chatApplyValue

	tests := []struct {
		name       string
		input      string
		targetFile string
		backupDir  string
		mockResp   string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "正常系：通常のチャット",
			input:      "Hello",
			targetFile: "",
			backupDir:  "",
			mockResp:   "Response from AI",
			mockErr:    nil,
			wantErr:    false,
		},
		{
			name:       "正常系：ファイル指定あり",
			input:      "コードを改善してください",
			targetFile: testFile,
			backupDir:  backupDir,
			mockResp: `---PROPOSAL---
コードを改善しました。
---CODE---
package main

func main() {
    fmt.Println("Hello")
}
---END---`,
			mockErr: nil,
			wantErr: false,
		},
		{
			name:       "異常系：ファイル読み込みエラー",
			input:      "Hello",
			targetFile: "nonexistent.go",
			backupDir:  backupDir,
			mockResp:   "",
			mockErr:    nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockAPIClient{
				response: tt.mockResp,
				err:      tt.mockErr,
			}

			result, err := executeChat(mockClient, tt.input, tt.targetFile, tt.backupDir)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}
