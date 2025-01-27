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
