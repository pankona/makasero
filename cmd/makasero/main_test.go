package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pankona/makasero/internal/models"
	"github.com/stretchr/testify/assert"
)

// モックAPIクライアント
type mockAPIClient struct {
	response string
	err      error
}

func (m *mockAPIClient) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
	return m.response, m.err
}

func TestParseAIResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		wantProposal  string
		wantCode      string
		wantErr       bool
		wantErrString string
	}{
		{
			name: "正常系：提案とコードを抽出",
			response: `---PROPOSAL---
エラーハンドリングを改善します。
---CODE---
func main() {
    fmt.Println("Hello")
}
---END---`,
			wantProposal: "エラーハンドリングを改善します。",
			wantCode: `func main() {
    fmt.Println("Hello")
}`,
			wantErr: false,
		},
		{
			name:          "異常系：不正なフォーマット",
			response:      "不正なレスポンス",
			wantProposal:  "",
			wantCode:      "",
			wantErr:       true,
			wantErrString: "不正なレスポンス形式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proposal, code, err := parseAIResponse(tt.response)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrString != "" {
					assert.Equal(t, tt.wantErrString, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantProposal, proposal)
				assert.Equal(t, tt.wantCode, code)
			}
		})
	}
}

func TestCreateBackup(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")

	// テスト用のファイルを作成
	testFile := filepath.Join(tmpDir, "test.go")
	testContent := []byte("package main\n\nfunc main() {}\n")
	err := os.WriteFile(testFile, testContent, 0644)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		filePath  string
		backupDir string
		wantErr   bool
	}{
		{
			name:      "正常系：バックアップを作成",
			filePath:  testFile,
			backupDir: backupDir,
			wantErr:   false,
		},
		{
			name:      "異常系：存在しないファイル",
			filePath:  filepath.Join(tmpDir, "notexist.go"),
			backupDir: backupDir,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createBackup(tt.filePath, tt.backupDir)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// バックアップディレクトリが作成されたことを確認
				_, err := os.Stat(tt.backupDir)
				assert.NoError(t, err)

				// バックアップファイルが作成されたことを確認
				files, err := os.ReadDir(tt.backupDir)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(files))

				// バックアップファイルの内容を確認
				backupPath := filepath.Join(tt.backupDir, files[0].Name())
				content, err := os.ReadFile(backupPath)
				assert.NoError(t, err)
				assert.Equal(t, testContent, content)
			}
		})
	}
}

func TestApplyChanges(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		filePath   string
		newContent string
		wantErr    bool
	}{
		{
			name:       "正常系：ファイルを更新",
			filePath:   filepath.Join(tmpDir, "test.go"),
			newContent: "package main\n\nfunc main() {}\n",
			wantErr:    false,
		},
		{
			name:       "異常系：書き込み権限なし",
			filePath:   "/root/test.go",
			newContent: "test",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyChanges(tt.filePath, tt.newContent)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// ファイルの内容を確認
				content, err := os.ReadFile(tt.filePath)
				assert.NoError(t, err)
				assert.Equal(t, tt.newContent, string(content))
			}
		})
	}
}

func TestExecuteChat(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	backupDir := filepath.Join(tmpDir, "backups")

	// テスト用のファイルを作成
	testContent := "package main\n\nfunc main() {}\n"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	assert.NoError(t, err)

	// テスト環境変数の設定
	os.Setenv("MAKASERO_TEST", "1")
	defer os.Unsetenv("MAKASERO_TEST")

	tests := []struct {
		name       string
		client     *mockAPIClient
		input      string
		targetFile string
		backupDir  string
		wantErr    bool
	}{
		{
			name: "正常系：ファイルなし",
			client: &mockAPIClient{
				response: "テスト応答",
				err:      nil,
			},
			input:   "テスト",
			wantErr: false,
		},
		{
			name: "正常系：ファイルあり",
			client: &mockAPIClient{
				response: `---PROPOSAL---
テスト提案
---CODE---
package main

func main() {
    fmt.Println("Hello")
}
---END---`,
				err: nil,
			},
			input:      "テスト",
			targetFile: testFile,
			backupDir:  backupDir,
			wantErr:    false,
		},
		{
			name: "異常系：APIエラー",
			client: &mockAPIClient{
				response: "",
				err:      assert.AnError,
			},
			input:   "テスト",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeChat(tt.client, tt.input, tt.targetFile, tt.backupDir)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteExplain(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testContent := "package main\n\nfunc main() {}\n"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		client  *mockAPIClient
		code    string
		wantErr bool
	}{
		{
			name: "正常系：コード文字列",
			client: &mockAPIClient{
				response: "コードの説明",
				err:      nil,
			},
			code:    "func main() {}",
			wantErr: false,
		},
		{
			name: "正常系：ファイルパス",
			client: &mockAPIClient{
				response: "ファイルの説明",
				err:      nil,
			},
			code:    testFile,
			wantErr: false,
		},
		{
			name: "異常系：APIエラー",
			client: &mockAPIClient{
				response: "",
				err:      assert.AnError,
			},
			code:    "func main() {}",
			wantErr: true,
		},
		{
			name: "異常系：存在しないファイル",
			client: &mockAPIClient{
				response: "",
				err:      nil,
			},
			code:    filepath.Join(tmpDir, "notexist.go"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeExplain(tt.client, tt.code)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	client := &mockAPIClient{
		response: "テスト応答",
		err:      nil,
	}

	tests := []struct {
		name       string
		command    string
		input      string
		targetFile string
		backupDir  string
		wantErr    bool
	}{
		{
			name:    "正常系：explainコマンド",
			command: "explain",
			input:   "func main() {}",
			wantErr: false,
		},
		{
			name:    "正常系：chatコマンド",
			command: "chat",
			input:   "テスト",
			wantErr: false,
		},
		{
			name:    "異常系：不明なコマンド",
			command: "unknown",
			input:   "テスト",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCommand(client, tt.command, tt.input, tt.targetFile, tt.backupDir)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
