package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleReadFile(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
		checkFn func(map[string]any) error
	}{
		{
			name: "正常系: ファイル全体を読み取る",
			args: map[string]any{
				"path": "testdata/test.txt",
			},
			wantErr: false,
			checkFn: func(result map[string]any) error {
				if !result["success"].(bool) {
					return fmt.Errorf("expected success=true, got false")
				}
				content := result["content"].(string)
				expected := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5 "
				if content != expected {
					return fmt.Errorf("unexpected content:\nwant: %q\ngot:  %q", expected, content)
				}
				return nil
			},
		},
		{
			name: "正常系: 特定の行範囲を読み取る",
			args: map[string]any{
				"path":       "testdata/test.txt",
				"start_line": float64(2),
				"end_line":   float64(4),
			},
			wantErr: false,
			checkFn: func(result map[string]any) error {
				if !result["success"].(bool) {
					return fmt.Errorf("expected success=true, got false")
				}
				content := result["content"].(string)
				expected := "Line 2\nLine 3\nLine 4"
				if content != expected {
					return fmt.Errorf("unexpected content:\nwant: %q\ngot:  %q", expected, content)
				}
				if result["start_line"].(int) != 2 {
					return fmt.Errorf("unexpected start_line: want 2, got %d", result["start_line"].(int))
				}
				if result["end_line"].(int) != 4 {
					return fmt.Errorf("unexpected end_line: want 4, got %d", result["end_line"].(int))
				}
				return nil
			},
		},
		{
			name: "異常系: 存在しないファイル",
			args: map[string]any{
				"path": "testdata/nonexistent.txt",
			},
			wantErr: false, // エラーは戻り値として返されるため、wantErrはfalse
			checkFn: func(result map[string]any) error {
				if result["success"].(bool) {
					return fmt.Errorf("expected success=false, got true")
				}
				if result["error"] == nil {
					return fmt.Errorf("expected error message, got nil")
				}
				return nil
			},
		},
		{
			name: "異常系: 無効な行範囲",
			args: map[string]any{
				"path":       "testdata/test.txt",
				"start_line": float64(10),
				"end_line":   float64(20),
			},
			wantErr: false,
			checkFn: func(result map[string]any) error {
				if result["success"].(bool) {
					return fmt.Errorf("expected success=false, got true")
				}
				if result["error"] == nil {
					return fmt.Errorf("expected error message, got nil")
				}
				return nil
			},
		},
		{
			name:    "異常系: 必須パラメータなし",
			args:    map[string]any{},
			wantErr: true,
			checkFn: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleReadFile(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFn != nil {
				if err := tt.checkFn(result); err != nil {
					t.Errorf("checkFn() error = %v", err)
				}
			}
		})
	}
}

func TestHandleWriteFile(t *testing.T) {
	ctx := context.Background()

	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
		checkFn func(map[string]any) error
	}{
		{
			name: "正常系: 新規ファイル作成",
			args: map[string]any{
				"path":    testFile,
				"content": "Hello, World!",
			},
			wantErr: false,
			checkFn: func(result map[string]any) error {
				if !result["success"].(bool) {
					return fmt.Errorf("expected success=true, got false")
				}
				// ファイルの内容を確認
				content, err := os.ReadFile(testFile)
				if err != nil {
					return err
				}
				if string(content) != "Hello, World!" {
					return fmt.Errorf("unexpected content: %s", string(content))
				}
				return nil
			},
		},
		{
			name: "正常系: 追記モード",
			args: map[string]any{
				"path":    testFile,
				"content": "\nGoodbye, World!",
				"append":  true,
			},
			wantErr: false,
			checkFn: func(result map[string]any) error {
				if !result["success"].(bool) {
					return fmt.Errorf("expected success=true, got false")
				}
				// ファイルの内容を確認
				content, err := os.ReadFile(testFile)
				if err != nil {
					return err
				}
				if string(content) != "Hello, World!\nGoodbye, World!" {
					return fmt.Errorf("unexpected content: %s", string(content))
				}
				return nil
			},
		},
		{
			name: "正常系: 特定の行に挿入",
			args: map[string]any{
				"path":       testFile,
				"content":    "Inserted line",
				"start_line": float64(2),
			},
			wantErr: false,
			checkFn: func(result map[string]any) error {
				if !result["success"].(bool) {
					return fmt.Errorf("expected success=true, got false")
				}
				// ファイルの内容を確認
				content, err := os.ReadFile(testFile)
				if err != nil {
					return err
				}
				expected := "Hello, World!\nInserted line\nGoodbye, World!"
				if string(content) != expected {
					return fmt.Errorf("unexpected content:\nwant: %q\ngot:  %q", expected, string(content))
				}
				return nil
			},
		},
		{
			name:    "異常系: 必須パラメータなし",
			args:    map[string]any{},
			wantErr: true,
			checkFn: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleWriteFile(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleWriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFn != nil {
				if err := tt.checkFn(result); err != nil {
					t.Errorf("checkFn() error = %v", err)
				}
			}
		})
	}
}
