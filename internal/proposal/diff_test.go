package proposal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffUtility_GenerateDiff(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		proposed  string
		wantEmpty bool
		wantError bool
	}{
		{
			name:      "正常系：単純な変更",
			original:  "Hello World",
			proposed:  "Hello Go",
			wantEmpty: false,
			wantError: false,
		},
		{
			name:      "正常系：変更なし",
			original:  "Hello World",
			proposed:  "Hello World",
			wantEmpty: true,
			wantError: false,
		},
		{
			name:      "正常系：空の文字列",
			original:  "",
			proposed:  "Hello",
			wantEmpty: false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDiffUtility()
			diff, err := d.GenerateDiff(tt.original, tt.proposed)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantEmpty {
					assert.Empty(t, diff)
				} else {
					assert.NotEmpty(t, diff)
				}
			}
		})
	}
}

func TestDiffUtility_ApplyPatch(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		proposed  string
		wantError bool
	}{
		{
			name:      "正常系：パッチの適用",
			original:  "Hello World",
			proposed:  "Hello Go",
			wantError: false,
		},
		{
			name:      "正常系：変更なし",
			original:  "Hello World",
			proposed:  "Hello World",
			wantError: false,
		},
		{
			name:      "正常系：空文字列からの変更",
			original:  "",
			proposed:  "Hello",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDiffUtility()

			// 差分の生成
			patch, err := d.GenerateDiff(tt.original, tt.proposed)
			assert.NoError(t, err)

			// パッチの適用
			result, err := d.ApplyPatch(tt.original, patch)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.proposed, result)
			}
		})
	}
}

func TestDiffUtility_FormatDiff(t *testing.T) {
	tests := []struct {
		name     string
		original string
		proposed string
		want     []string
	}{
		{
			name:     "正常系：削除と追加",
			original: "Hello World",
			proposed: "Hello Go",
			want: []string{
				"- W",
				"+ G",
				"- rld",
			},
		},
		{
			name:     "正常系：追加のみ",
			original: "Hello",
			proposed: "Hello World",
			want: []string{
				"+ World",
			},
		},
		{
			name:     "正常系：削除のみ",
			original: "Hello World",
			proposed: "Hello",
			want: []string{
				"- World",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDiffUtility()
			formatted := d.FormatDiff(tt.original, tt.proposed)

			for _, line := range tt.want {
				assert.True(t, strings.Contains(formatted, line),
					"expected to find line %q in formatted output: %s", line, formatted)
			}
		})
	}
}

func TestSearchReplaceDiffStrategy_ApplyDiff(t *testing.T) {
	tests := []struct {
		name           string
		original       string
		searchContent  string
		replaceContent string
		startLine      int
		endLine        int
		fuzzyThreshold float64
		wantError      bool
		errorType      string
		want           string
	}{
		{
			name: "正常系：完全一致での置換",
			original: `func hello() {
    fmt.Println("Hello")
}`,
			searchContent:  `    fmt.Println("Hello")`,
			replaceContent: `    fmt.Println("Hello, World!")`,
			startLine:      2,
			endLine:        2,
			fuzzyThreshold: 1.0,
			wantError:      false,
			want: `func hello() {
    fmt.Println("Hello, World!")
}`,
		},
		{
			name: "異常系：類似度不足",
			original: `func hello() {
    fmt.Printf("Hello")
}`,
			searchContent:  `    fmt.Println("Hello")`,
			replaceContent: `    fmt.Println("Hello, World!")`,
			startLine:      2,
			endLine:        2,
			fuzzyThreshold: 1.0,
			wantError:      true,
			errorType:      "DiffError",
		},
		{
			name: "異常系：行範囲外",
			original: `func hello() {
    fmt.Println("Hello")
}`,
			searchContent:  `    fmt.Println("Hello")`,
			replaceContent: `    fmt.Println("Hello, World!")`,
			startLine:      5,
			endLine:        5,
			fuzzyThreshold: 1.0,
			wantError:      true,
			errorType:      "DiffError",
		},
		{
			name: "異常系：空の検索内容",
			original: `func hello() {
    fmt.Println("Hello")
}`,
			searchContent:  ``,
			replaceContent: `    fmt.Println("Hello, World!")`,
			startLine:      0,
			endLine:        0,
			fuzzyThreshold: 1.0,
			wantError:      true,
			errorType:      "DiffError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSearchReplaceDiffStrategy(tt.fuzzyThreshold, DefaultBufferLines)
			result, err := s.ApplyDiff(tt.original, tt.searchContent, tt.replaceContent, tt.startLine, tt.endLine)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.IsType(t, &DiffError{}, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestGetSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		original string
		search   string
		want     float64
	}{
		{
			name:     "正常系：完全一致",
			original: "Hello, World!",
			search:   "Hello, World!",
			want:     1.0,
		},
		{
			name:     "正常系：部分一致",
			original: "Hello, World!",
			search:   "Hello",
			want:     0.3846153846153846,
		},
		{
			name:     "正常系：大文字小文字の違い",
			original: "Hello, World!",
			search:   "hello, world!",
			want:     0.8461538461538461,
		},
		{
			name:     "正常系：空文字列",
			original: "",
			search:   "",
			want:     1.0,
		},
		{
			name:     "正常系：空白の違い",
			original: "Hello,  World!",
			search:   "Hello, World!",
			want:     1.0,
		},
		{
			name:     "正常系：コメントの除去",
			original: "Hello, World! // コメント",
			search:   "Hello, World!",
			want:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSimilarity(tt.original, tt.search)
			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestDiffError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *DiffError
		want string
	}{
		{
			name: "正常系：すべてのフィールドが設定されている場合",
			err: &DiffError{
				Message:     "テストエラー",
				Context:     "検索コンテキスト",
				Similarity:  0.8,
				RequiredSim: 0.9,
				SearchRange: "1-10",
				BestMatch:   "最も近い一致",
				OrigContent: "元のコンテンツ",
			},
			want: `テストエラー

デバッグ情報:
- 類似度スコア: 80%
- 必要な類似度: 90%
- 検索範囲: 1-10

検索内容:
検索コンテキスト

最も近い一致:
最も近い一致

元のコンテンツ:
元のコンテンツ`,
		},
		{
			name: "正常系：最小限のフィールドのみ設定されている場合",
			err: &DiffError{
				Message:     "最小限のエラー",
				Similarity:  0.5,
				RequiredSim: 1.0,
			},
			want: `最小限のエラー

デバッグ情報:
- 類似度スコア: 50%
- 必要な類似度: 100%
- 検索範囲: 

検索内容:


最も近い一致:


元のコンテンツ:
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		want    int
	}{
		{
			name:    "正常系：正の数",
			numbers: []int{3, 1, 4, 1, 5},
			want:    1,
		},
		{
			name:    "正常系：負の数を含む",
			numbers: []int{-1, 2, -3, 4},
			want:    -3,
		},
		{
			name:    "正常系：同じ数を含む",
			numbers: []int{2, 2, 2},
			want:    2,
		},
		{
			name:    "正常系：1つの数",
			numbers: []int{42},
			want:    42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := min(tt.numbers...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{
			name: "正常系：aがbより大きい",
			a:    5,
			b:    3,
			want: 5,
		},
		{
			name: "正常系：bがaより大きい",
			a:    2,
			b:    4,
			want: 4,
		},
		{
			name: "正常系：aとbが等しい",
			a:    7,
			b:    7,
			want: 7,
		},
		{
			name: "正常系：負の数",
			a:    -2,
			b:    -5,
			want: -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := max(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
