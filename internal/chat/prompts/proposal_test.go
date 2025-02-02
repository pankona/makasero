package prompts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProposalSystemPrompt(t *testing.T) {
	tests := []struct {
		name       string
		checkPoint string
	}{
		{
			name:       "必須セクション：PROPOSAL",
			checkPoint: "---PROPOSAL---",
		},
		{
			name:       "必須セクション：FILE",
			checkPoint: "---FILE---",
		},
		{
			name:       "必須セクション：DIFF",
			checkPoint: "---DIFF---",
		},
		{
			name:       "必須セクション：END",
			checkPoint: "---END---",
		},
		{
			name:       "Unified diff形式の説明",
			checkPoint: "@@ -1,3 +1,4 @@",
		},
		{
			name:       "改善の観点：コードの品質",
			checkPoint: "1. コードの品質向上",
		},
		{
			name:       "改善の観点：エラーハンドリング",
			checkPoint: "2. エラーハンドリングの改善",
		},
		{
			name:       "改善の観点：パフォーマンス",
			checkPoint: "3. パフォーマンスの最適化",
		},
		{
			name:       "改善の観点：可読性",
			checkPoint: "4. 可読性の向上",
		},
		{
			name:       "改善の観点：ベストプラクティス",
			checkPoint: "5. ベストプラクティスの適用",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, ProposalSystemPrompt, tt.checkPoint)
		})
	}
}

func TestProposalMarkers(t *testing.T) {
	tests := []struct {
		name   string
		marker Markers
	}{
		{
			name: "正常系：デフォルトマーカー",
			marker: Markers{
				Proposal: "---PROPOSAL---",
				File:     "---FILE---",
				Diff:     "---DIFF---",
				End:      "---END---",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.marker, ProposalMarkers)
		})
	}
}

func TestProposalSystemPrompt_Integration(t *testing.T) {
	// プロンプトにマーカーが含まれていることを確認
	assert.True(t, strings.Contains(ProposalSystemPrompt, ProposalMarkers.Proposal))
	assert.True(t, strings.Contains(ProposalSystemPrompt, ProposalMarkers.File))
	assert.True(t, strings.Contains(ProposalSystemPrompt, ProposalMarkers.Diff))
	assert.True(t, strings.Contains(ProposalSystemPrompt, ProposalMarkers.End))

	// プロンプトの形式が正しいことを確認
	lines := strings.Split(ProposalSystemPrompt, "\n")
	var hasCodeBlock bool
	for _, line := range lines {
		if strings.Contains(line, "```go") {
			hasCodeBlock = true
			break
		}
	}
	assert.True(t, hasCodeBlock, "プロンプトにGoのコードブロックが含まれていること")
}
