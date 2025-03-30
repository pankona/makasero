package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile_Execute(t *testing.T) {
	// テスト用のディレクトリを作成
	tmpDir := t.TempDir()

	// テストファイルを作成
	testContent := "Hello, World!"
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "正常なファイル読み取り",
			path:    testFile,
			want:    testContent,
			wantErr: false,
		},
		{
			name:    "存在しないファイル",
			path:    filepath.Join(tmpDir, "nonexistent.txt"),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReadFile{
				Path: tt.path,
			}

			got, err := r.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
} 