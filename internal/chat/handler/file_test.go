package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pankona/makasero/internal/chat/detector"
)

func TestFileApplier_Apply(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "file-applier-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// バックアップディレクトリの設定
	backupDir := filepath.Join(tempDir, "backups")

	// テストケース用のファイルを作成
	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "Hello\nWorld\n"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		diff        string
		wantContent string
		wantErr     bool
	}{
		{
			name:     "正常なパッチ適用",
			filePath: testFile,
			diff: `--- test.txt
+++ test.txt
@@ -1,2 +1,2 @@
 Hello
-World
+Roo
`,
			wantContent: "Hello\nRoo\n",
			wantErr:     false,
		},
		{
			name:     "存在しないファイル",
			filePath: filepath.Join(tempDir, "nonexistent.txt"),
			diff:     "",
			wantErr:  true,
		},
		{
			name:     "不正なパッチ",
			filePath: testFile,
			diff: `--- test.txt
+++ test.txt
@@ -1,2 +1,2 @@
-This line does not exist
+New line
`,
			wantContent: originalContent, // バックアップから復元されるべき
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストケースごとにファイルを元の状態に戻す
			if tt.filePath == testFile {
				if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
					t.Fatalf("ファイルの初期化に失敗: %v", err)
				}
			}

			applier, err := NewFileApplier(backupDir)
			if err != nil {
				t.Fatalf("FileApplierの作成に失敗: %v", err)
			}

			proposal := &detector.Proposal{
				FilePath: tt.filePath,
				Diff:     tt.diff,
			}

			err = applier.Apply(proposal)
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.filePath == testFile {
				// 結果の確認
				content, err := os.ReadFile(testFile)
				if err != nil {
					t.Fatalf("結果の読み取りに失敗: %v", err)
				}
				if got := string(content); got != tt.wantContent {
					t.Errorf("ファイルの内容が一致しません\ngot:\n%s\nwant:\n%s", got, tt.wantContent)
				}
			}

			// バックアップの確認
			if tt.filePath == testFile {
				files, err := os.ReadDir(backupDir)
				if err != nil {
					t.Fatalf("バックアップディレクトリの読み取りに失敗: %v", err)
				}
				found := false
				for _, f := range files {
					if strings.HasPrefix(f.Name(), "test.txt.") && strings.HasSuffix(f.Name(), ".bak") {
						found = true
						break
					}
				}
				if !found {
					t.Error("バックアップファイルが作成されていません")
				}
			}
		})
	}
}
