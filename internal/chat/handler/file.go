package handler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pankona/makasero/internal/chat/detector"
)

// FileApplier は、ファイルへの変更を適用する実装です。
type FileApplier struct {
	backupDir string
}

// NewFileApplier は、新しいFileApplierインスタンスを作成します。
func NewFileApplier(backupDir string) (*FileApplier, error) {
	// バックアップディレクトリの作成
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("バックアップディレクトリの作成に失敗しました: %w", err)
	}

	return &FileApplier{
		backupDir: backupDir,
	}, nil
}

// Apply は、提案の変更をファイルに適用します。
func (a *FileApplier) Apply(proposal *detector.Proposal) error {
	// 1. ファイルの存在確認
	if _, err := os.Stat(proposal.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("対象ファイルが存在しません: %s", proposal.FilePath)
	}

	// 2. バックアップの作成
	backupPath, err := a.createBackup(proposal.FilePath)
	if err != nil {
		return fmt.Errorf("バックアップの作成に失敗しました: %w", err)
	}

	// 3. パッチファイルの作成
	patchFile, err := os.CreateTemp("", "proposal-*.patch")
	if err != nil {
		return fmt.Errorf("パッチファイルの作成に失敗しました: %w", err)
	}
	defer os.Remove(patchFile.Name())

	if _, err := patchFile.WriteString(proposal.Diff); err != nil {
		return fmt.Errorf("パッチの書き込みに失敗しました: %w", err)
	}
	patchFile.Close()

	// 4. パッチの適用
	cmd := exec.Command("patch", proposal.FilePath, patchFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		// エラー時はバックアップから復元
		if restoreErr := a.restoreFromBackup(backupPath, proposal.FilePath); restoreErr != nil {
			return fmt.Errorf("パッチの適用に失敗し、バックアップからの復元にも失敗しました: %v (元のエラー: %v)\nパッチ出力: %s", restoreErr, err, output)
		}
		return fmt.Errorf("パッチの適用に失敗しました: %w\nパッチ出力: %s", err, output)
	}

	return nil
}

// createBackup は、ファイルのバックアップを作成します。
func (a *FileApplier) createBackup(filePath string) (string, error) {
	backupPath := filepath.Join(a.backupDir,
		fmt.Sprintf("%s.%d.bak",
			filepath.Base(filePath),
			time.Now().UnixNano()))

	if err := copyFile(filePath, backupPath); err != nil {
		return "", err
	}

	return backupPath, nil
}

// restoreFromBackup は、バックアップからファイルを復元します。
func (a *FileApplier) restoreFromBackup(backupPath, originalPath string) error {
	return copyFile(backupPath, originalPath)
}

// copyFile は、ファイルをコピーします。
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}
