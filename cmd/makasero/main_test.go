package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pankona/makasero/internal/models"
	"github.com/stretchr/testify/assert"
)

// osExitをモック化するための変数
var osExit = os.Exit

// エラーハンドリング用のヘルパー関数
func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		osExit(1)
	}
	osExit(0)
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

func TestOutputResponse(t *testing.T) {
	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w

	// テスト終了時に元の標準出力を復元
	defer func() {
		os.Stdout = oldStdout
		r.Close()
		w.Close()
	}()

	// テストレスポンスを出力
	response := models.Response{
		Success: true,
		Data:    "テストレスポンス",
	}
	outputResponse(response)
	w.Close()

	// 出力を読み取り
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	// JSONとして解析可能か確認
	var got models.Response
	err = json.Unmarshal(buf.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, response.Success, got.Success)
	assert.Equal(t, response.Data, got.Data)
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantExit int
	}{
		{
			name:     "正常系：エラーなし",
			err:      nil,
			wantExit: 0,
		},
		{
			name:     "異常系：エラーあり",
			err:      fmt.Errorf("テストエラー"),
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準エラー出力をキャプチャ
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// exit関数をモック
			var exitCode int
			oldOsExit := osExit
			osExit = func(code int) {
				exitCode = code
				panic("test exit") // テスト用にpanicを発生させる
			}

			// テスト終了時に元の状態を復元
			defer func() {
				os.Stderr = oldStderr
				osExit = oldOsExit
				r.Close()
				w.Close()
				// panicをリカバー
				recover()
			}()

			handleError(tt.err)
			w.Close()

			// 出力を読み取り
			var buf bytes.Buffer
			io.Copy(&buf, r)

			if tt.err != nil {
				assert.Contains(t, buf.String(), tt.err.Error())
			}
			assert.Equal(t, tt.wantExit, exitCode)
		})
	}
}
