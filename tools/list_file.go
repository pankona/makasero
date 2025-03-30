package tools

import (
	"os"
	"path/filepath"
	"strings"
)

// ListFile はディレクトリ内のファイル一覧を取得します
type ListFile struct {
	Path      string
	Recursive bool
}

// Execute はファイル一覧を取得して返します
func (l *ListFile) Execute() (string, error) {
	var files []string
	var err error

	if l.Recursive {
		err = filepath.Walk(l.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(l.Path)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(l.Path, entry.Name()))
			}
		}
	}

	if err != nil {
		return "", err
	}

	return strings.Join(files, "\n"), nil
}
