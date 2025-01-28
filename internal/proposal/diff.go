package proposal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	// DefaultFuzzyThreshold はデフォルトの類似度閾値です（100%の一致を要求）
	DefaultFuzzyThreshold = 1.0
	// DefaultBufferLines はデフォルトのコンテキスト行数です
	DefaultBufferLines = 20
)

// DiffError は差分適用時のエラーを表します
type DiffError struct {
	Message     string
	Context     string
	Similarity  float64
	RequiredSim float64
	SearchRange string
	BestMatch   string
	OrigContent string
}

func (e *DiffError) Error() string {
	return fmt.Sprintf("%s\n\nデバッグ情報:\n- 類似度スコア: %.0f%%\n- 必要な類似度: %.0f%%\n- 検索範囲: %s\n\n検索内容:\n%s\n\n最も近い一致:\n%s\n\n元のコンテンツ:\n%s",
		e.Message,
		e.Similarity*100,
		e.RequiredSim*100,
		e.SearchRange,
		e.Context,
		e.BestMatch,
		e.OrigContent,
	)
}

// DiffStrategy は差分適用の戦略を定義するインターフェースです
type DiffStrategy interface {
	// ApplyDiff は差分を適用して新しいテキストを生成します
	ApplyDiff(original, searchContent, replaceContent string, startLine, endLine int) (string, error)
}

// SearchReplaceDiffStrategy は検索と置換に基づく差分適用戦略を提供します
type SearchReplaceDiffStrategy struct {
	fuzzyThreshold float64
	bufferLines    int
}

// NewSearchReplaceDiffStrategy は新しいSearchReplaceDiffStrategyインスタンスを作成します
func NewSearchReplaceDiffStrategy(fuzzyThreshold float64, bufferLines int) *SearchReplaceDiffStrategy {
	return &SearchReplaceDiffStrategy{
		fuzzyThreshold: fuzzyThreshold,
		bufferLines:    bufferLines,
	}
}

// levenshteinDistance は2つの文字列間のLevenshtein距離を計算します
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j-1]+1, // 置換
					matrix[i][j-1]+1,   // 挿入
					matrix[i-1][j]+1,   // 削除
				)
			}
		}
	}

	return matrix[len(a)][len(b)]
}

// min は与えられた整数の最小値を返します
func min(numbers ...int) int {
	if len(numbers) == 0 {
		return 0
	}
	min := numbers[0]
	for _, num := range numbers[1:] {
		if num < min {
			min = num
		}
	}
	return min
}

// max は2つの整数の最大値を返します
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getSimilarity は2つの文字列間の類似度を計算します
func getSimilarity(original, search string) float64 {
	if search == "" {
		return 1
	}

	// コメントを除去
	removeComment := func(str string) string {
		if idx := strings.Index(str, "//"); idx != -1 {
			return strings.TrimSpace(str[:idx])
		}
		return str
	}

	// 文字列を正規化（余分な空白を削除）
	normalizeStr := func(str string) string {
		// コメントを除去してから正規化
		str = removeComment(str)
		return strings.Join(strings.Fields(str), " ")
	}

	normalizedOriginal := normalizeStr(original)
	normalizedSearch := normalizeStr(search)

	if normalizedOriginal == normalizedSearch {
		return 1
	}

	// Levenshtein距離を計算
	distance := levenshteinDistance(normalizedOriginal, normalizedSearch)

	// 類似度を計算（0から1の範囲、1が完全一致）
	maxLength := max(len(normalizedOriginal), len(normalizedSearch))
	if maxLength == 0 {
		return 1
	}
	return 1 - float64(distance)/float64(maxLength)
}

// preserveIndentation は元のインデントを保持しながら置換を行います
func (s *SearchReplaceDiffStrategy) preserveIndentation(originalLines, searchLines, replaceLines []string) []string {
	if len(originalLines) == 0 || len(searchLines) == 0 {
		return replaceLines
	}

	// 元のインデントを取得
	getIndent := func(line string) string {
		matches := regexp.MustCompile(`^[ \t]*`).FindString(line)
		return matches
	}

	// インデントパターンを解析
	analyzeIndentPattern := func(line string) (indent string, spaceCount int, isSpaceBased bool) {
		indent = getIndent(line)
		if len(indent) == 0 {
			return "", 0, false
		}

		// スペースベースのインデントを検出
		if !strings.Contains(indent, "\t") {
			// スペースのみの場合、元のインデントを保持
			return indent, 0, true
		}

		// タブベースのインデントを解析
		parts := strings.Split(indent, "\t")
		tabCount := len(parts) - 1
		baseIndent := strings.Repeat("\t", tabCount)

		// タブ後のスペースを計算
		if len(parts) > 1 {
			lastPart := parts[len(parts)-1]
			// タブ後のスペースを正確に計算
			spaceCount = len(strings.TrimRight(lastPart, "\t"))
			if spaceCount > 0 {
				// スペースの数を4の倍数に調整
				spaceCount = ((spaceCount + 3) / 4) * 4
			}
		}

		return baseIndent, spaceCount, false
	}

	// インデントレベルを計算
	getIndentLevel := func(line string) int {
		indent := getIndent(line)
		if strings.Contains(indent, "\t") {
			return strings.Count(indent, "\t")
		}
		return len(indent) / 4
	}

	// 行のインデントを調整
	adjustIndent := func(line string) string {
		if strings.TrimSpace(line) == "" {
			return line // 空行はそのまま
		}

		// 元のインデントパターンを解析
		origIndent, spaceCount, isSpaceBased := analyzeIndentPattern(originalLines[0])

		// 現在の行のインデントレベルを取得
		currentLevel := getIndentLevel(line)
		baseLevel := getIndentLevel(searchLines[0])
		targetLevel := getIndentLevel(originalLines[0])

		// 相対的なインデントレベルを計算
		relativeLevel := currentLevel - baseLevel
		finalLevel := targetLevel + relativeLevel

		// インデントを構築
		var finalIndent string
		if !isSpaceBased {
			// タブベースのインデント
			finalIndent = strings.Repeat("\t", finalLevel)
			if spaceCount > 0 {
				// タブ後のスペースを計算
				spaceIndent := strings.Repeat(" ", spaceCount)
				// 相対レベルに応じてスペースを追加
				if relativeLevel > 0 {
					// 元のスペース数を基準に増分を計算
					additionalSpaces := spaceCount * relativeLevel
					spaceIndent = strings.Repeat(" ", spaceCount+additionalSpaces)
				}
				finalIndent += spaceIndent
			}
		} else {
			// スペースベースのインデント
			if len(origIndent) > 0 {
				// 元のインデントパターンを使用
				finalIndent = strings.Repeat(origIndent, finalLevel)
			} else {
				// デフォルトの4スペースを使用
				finalIndent = strings.Repeat("    ", finalLevel)
			}
		}

		// 行の内容を追加
		return finalIndent + strings.TrimLeft(line, " \t")
	}

	// 置換する行を処理
	result := make([]string, len(replaceLines))
	for i, line := range replaceLines {
		result[i] = adjustIndent(line)
	}

	return result
}

// ApplyDiff は差分を適用して新しいテキストを生成します
func (s *SearchReplaceDiffStrategy) ApplyDiff(original, searchContent, replaceContent string, startLine, endLine int) (string, error) {
	// 行を分割
	originalLines := strings.Split(original, "\n")
	searchLines := strings.Split(searchContent, "\n")
	replaceLines := strings.Split(replaceContent, "\n")

	// 空の検索内容の場合は開始行が必要
	if len(searchLines) == 0 && startLine == 0 {
		return "", &DiffError{
			Message: "空の検索内容には開始行の指定が必要です",
			Context: "挿入操作には特定の行番号が必要です",
		}
	}

	// 空の検索内容の場合は開始行と終了行が同じである必要がある
	if len(searchLines) == 0 && startLine != endLine {
		return "", &DiffError{
			Message: fmt.Sprintf("空の検索内容には同じ開始行と終了行が必要です（指定: %d-%d）", startLine, endLine),
			Context: "挿入操作には同じ行番号を指定してください",
		}
	}

	// 行範囲の検証
	if startLine > 0 && endLine > 0 {
		// 0ベースのインデックスに変換
		exactStartIndex := startLine - 1
		exactEndIndex := endLine - 1

		if exactStartIndex < 0 || exactEndIndex >= len(originalLines) || exactStartIndex > exactEndIndex {
			return "", &DiffError{
				Message:     fmt.Sprintf("行範囲 %d-%d が不正です（ファイルは %d 行）", startLine, endLine, len(originalLines)),
				SearchRange: fmt.Sprintf("行 %d-%d", startLine, endLine),
			}
		}

		// コメントを抽出
		getComment := func(line string) string {
			if idx := strings.Index(line, "//"); idx != -1 {
				return line[idx:]
			}
			return ""
		}

		// 指定された範囲で正確な一致を試みる
		originalChunk := strings.Join(originalLines[exactStartIndex:exactEndIndex+1], "\n")
		similarity := getSimilarity(originalChunk, searchContent)

		if similarity >= s.fuzzyThreshold {
			// 一致が見つかった場合、置換を実行
			beforeMatch := originalLines[:exactStartIndex]
			afterMatch := originalLines[exactEndIndex+1:]

			// コメントを保持して置換
			originalComments := make([]string, len(originalLines[exactStartIndex:exactEndIndex+1]))
			for i, line := range originalLines[exactStartIndex : exactEndIndex+1] {
				originalComments[i] = getComment(line)
			}

			// インデントを保持して置換
			matchedLines := originalLines[exactStartIndex : exactEndIndex+1]
			indentedReplaceLines := s.preserveIndentation(matchedLines, searchLines, replaceLines)

			// コメントを復元
			for i := range indentedReplaceLines {
				if i < len(originalComments) && originalComments[i] != "" {
					indentedReplaceLines[i] += "  " + originalComments[i]
				}
			}

			// 結果を結合
			result := make([]string, 0, len(beforeMatch)+len(indentedReplaceLines)+len(afterMatch))
			result = append(result, beforeMatch...)
			result = append(result, indentedReplaceLines...)
			result = append(result, afterMatch...)

			return strings.Join(result, "\n"), nil
		}

		// 一致しない場合、バッファ付きの範囲で検索
		searchStartIndex := max(0, startLine-s.bufferLines-1)
		searchEndIndex := min(len(originalLines), endLine+s.bufferLines)

		// バッファ範囲内で最も類似度の高い部分を探す
		bestMatch := ""
		bestMatchScore := similarity
		bestMatchIndex := exactStartIndex

		for i := searchStartIndex; i <= searchEndIndex-len(searchLines); i++ {
			chunk := strings.Join(originalLines[i:i+len(searchLines)], "\n")
			currentSimilarity := getSimilarity(chunk, searchContent)

			if currentSimilarity > bestMatchScore {
				bestMatchScore = currentSimilarity
				bestMatch = chunk
				bestMatchIndex = i
			}
		}

		if bestMatchScore >= s.fuzzyThreshold {
			// 最良の一致が見つかった場合、置換を実行
			beforeMatch := originalLines[:bestMatchIndex]
			afterMatch := originalLines[bestMatchIndex+len(searchLines):]

			// コメントを保持して置換
			originalComments := make([]string, len(originalLines[bestMatchIndex:bestMatchIndex+len(searchLines)]))
			for i, line := range originalLines[bestMatchIndex : bestMatchIndex+len(searchLines)] {
				originalComments[i] = getComment(line)
			}

			// インデントを保持して置換
			matchedLines := originalLines[bestMatchIndex : bestMatchIndex+len(searchLines)]
			indentedReplaceLines := s.preserveIndentation(matchedLines, searchLines, replaceLines)

			// コメントを復元
			for i := range indentedReplaceLines {
				if i < len(originalComments) && originalComments[i] != "" {
					indentedReplaceLines[i] += "  " + originalComments[i]
				}
			}

			// 結果を結合
			result := make([]string, 0, len(beforeMatch)+len(indentedReplaceLines)+len(afterMatch))
			result = append(result, beforeMatch...)
			result = append(result, indentedReplaceLines...)
			result = append(result, afterMatch...)

			return strings.Join(result, "\n"), nil
		}

		// 十分な一致が見つからなかった場合
		return "", &DiffError{
			Message:     fmt.Sprintf("十分な類似の一致が見つかりません（類似度 %.0f%%, 必要 %.0f%%）", bestMatchScore*100, s.fuzzyThreshold*100),
			Similarity:  bestMatchScore,
			RequiredSim: s.fuzzyThreshold,
			SearchRange: fmt.Sprintf("行 %d-%d", startLine, endLine),
			Context:     searchContent,
			BestMatch:   bestMatch,
			OrigContent: strings.Join(originalLines[searchStartIndex:searchEndIndex], "\n"),
		}
	}

	// 行範囲が指定されていない場合は全体を検索
	return "", &DiffError{
		Message: "行範囲の指定が必要です",
		Context: searchContent,
	}
}

// DiffUtility は差分の生成と適用を行うユーティリティです
type DiffUtility struct {
	dmp      *diffmatchpatch.DiffMatchPatch
	strategy DiffStrategy
}

// NewDiffUtility は新しいDiffUtilityインスタンスを作成します
func NewDiffUtility(fuzzyThreshold ...float64) *DiffUtility {
	threshold := DefaultFuzzyThreshold
	if len(fuzzyThreshold) > 0 {
		threshold = fuzzyThreshold[0]
	}
	return &DiffUtility{
		dmp:      diffmatchpatch.New(),
		strategy: NewSearchReplaceDiffStrategy(threshold, DefaultBufferLines),
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

	// コンフリクトマーカーを除去する正規表現（厳密なバージョン）
	conflictMarkerRegex := regexp.MustCompile(`(?m)^[<>]{7}.*\n(?:.+\n)*?={7}\n(?:.+\n)*?>{7}.*`)
	newText, applied := d.dmp.PatchApply(patches, original)

	// コンフリクトマーカーを除去（複数回適用）
	cleanedText := conflictMarkerRegex.ReplaceAllString(newText, "")
	cleanedText = conflictMarkerRegex.ReplaceAllString(cleanedText, "")

	// コメント内のCONFLICT表記を全て除去
	finalText := regexp.MustCompile(`//\s*CONFLICT.*`).ReplaceAllString(cleanedText, "")

	// パッチの適用結果を確認
	for _, success := range applied {
		if !success {
			return "", fmt.Errorf("パッチの適用に失敗しました")
		}
	}

	// 最終的なクリーンなテキストを返す
	return finalText, nil
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
