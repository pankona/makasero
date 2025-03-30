package tools

import (
	"os"
	"path/filepath"
)

// ReadFile はファイルの内容を読み取ります
type ReadFile struct {
	Path string
}

// Execute はファイルの内容を読み取って返します
func (r *ReadFile) Execute() (string, error) {
	// パスの検証
	absPath, err := filepath.Abs(r.Path)
	if err != nil {
		return "", err
	}

	// ファイルの存在確認
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", err
	}

	// ファイルの読み取り
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
} 