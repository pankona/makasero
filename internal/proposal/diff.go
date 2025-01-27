package proposal

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffUtility は差分の生成と適用を行うユーティリティです
type DiffUtility struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

// NewDiffUtility は新しいDiffUtilityインスタンスを作成します
func NewDiffUtility() *DiffUtility {
	return &DiffUtility{
		dmp: diffmatchpatch.New(),
	}
}

// GenerateDiff は2つのテキスト間の差分を生成します
func (d *DiffUtility) GenerateDiff(original, proposed string) (string, error) {
	diffs := d.dmp.DiffMain(original, proposed, true)

	// 差分がない場合は空の差分を返す
	if len(diffs) == 1 && diffs[0].Type == diffmatchpatch.DiffEqual {
		return "", nil
	}

	// 差分をUnified形式で出力
	patches := d.dmp.PatchMake(original, diffs)
	return d.dmp.PatchToText(patches), nil
}

// ApplyPatch はパッチを適用して新しいテキストを生成します
func (d *DiffUtility) ApplyPatch(original, patchText string) (string, error) {
	patches, err := d.dmp.PatchFromText(patchText)
	if err != nil {
		return "", fmt.Errorf("パッチの解析に失敗: %w", err)
	}

	if patchText == "" {
		return original, nil
	}

	newText, applied := d.dmp.PatchApply(patches, original)

	// パッチの適用結果を確認
	for _, success := range applied {
		if !success {
			return "", fmt.Errorf("パッチの適用に失敗しました")
		}
	}

	return newText, nil
}

// FormatDiff は差分を人間が読みやすい形式にフォーマットします
func (d *DiffUtility) FormatDiff(original, proposed string) string {
	diffs := d.dmp.DiffMain(original, proposed, true)
	var result strings.Builder
	result.WriteString("変更内容:\n\n")

	// 行ごとの差分を生成
	lines := []string{}

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffDelete:
			lines = append(lines, fmt.Sprintf("- %s", strings.TrimSpace(diff.Text)))
		case diffmatchpatch.DiffInsert:
			lines = append(lines, fmt.Sprintf("+ %s", strings.TrimSpace(diff.Text)))
		case diffmatchpatch.DiffEqual:
			text := strings.TrimSpace(diff.Text)
			if text != "" {
				lines = append(lines, fmt.Sprintf("  %s", text))
			}
		}
	}

	result.WriteString(strings.Join(lines, "\n"))
	result.WriteString("\n")

	return result.String()
}
