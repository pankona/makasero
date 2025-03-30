package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListFile_Execute(t *testing.T) {
	// テスト用のディレクトリを作成
	tmpDir := t.TempDir()

	// テストファイルを作成
	files := []string{
		"test1.txt",
		"test2.txt",
		"subdir/test3.txt",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		err = os.WriteFile(path, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tests := []struct {
		name      string
		path      string
		recursive bool
		want      int // 期待されるファイル数
	}{
		{
			name:      "非再帰的な一覧",
			path:      tmpDir,
			recursive: false,
			want:      2, // test1.txt, test2.txt
		},
		{
			name:      "再帰的な一覧",
			path:      tmpDir,
			recursive: true,
			want:      3, // test1.txt, test2.txt, subdir/test3.txt
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &ListFile{
				Path:      tt.path,
				Recursive: tt.recursive,
			}

			got, err := l.Execute()
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			// 結果を改行で分割してファイル数をカウント
			fileCount := len(strings.Split(strings.TrimSpace(got), "\n"))
			if fileCount != tt.want {
				t.Errorf("Execute() returned %d files, want %d", fileCount, tt.want)
			}
		})
	}
}
