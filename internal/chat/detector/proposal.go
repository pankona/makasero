package detector

import (
	"strings"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/chat/prompts"
)

// ProposalDetector は、AIの応答から提案を検出し抽出する機能を提供します。
type ProposalDetector struct {
	markers prompts.Markers
}

// NewProposalDetector は、新しいProposalDetectorインスタンスを作成します。
func NewProposalDetector() *ProposalDetector {
	return &ProposalDetector{
		markers: prompts.ProposalMarkers,
	}
}

// Proposal は、検出された提案の情報を保持する構造体です。
type Proposal struct {
	Description string // 提案の説明
	FilePath    string // 対象ファイルパス
	Diff        string // 変更内容
}

// IsProposal は、与えられた応答が提案を含むかどうかを判定します。
func (d *ProposalDetector) IsProposal(response string) bool {
	return strings.Contains(response, d.markers.Proposal)
}

// Extract は、応答から提案情報を抽出します。
// 提案が含まれていない場合は、nilを返します。
func (d *ProposalDetector) Extract(response string) (*Proposal, error) {
	if !d.IsProposal(response) {
		return nil, nil
	}

	// 最初の提案のみを抽出（重複を防ぐ）
	firstProposalStart := strings.Index(response, d.markers.Proposal)
	firstProposalEnd := strings.Index(response[firstProposalStart:], d.markers.End)
	if firstProposalEnd == -1 {
		return nil, ErrInvalidFormat
	}
	response = response[firstProposalStart : firstProposalStart+firstProposalEnd+len(d.markers.End)]

	// 各セクションの抽出
	description := extractSection(response, d.markers.Proposal, d.markers.File)
	filePath := extractSection(response, d.markers.File, d.markers.Diff)
	diff := extractSection(response, d.markers.Diff, d.markers.End)

	// 必須フィールドの検証
	if description == "" || filePath == "" || diff == "" {
		return nil, ErrInvalidFormat
	}

	// コードブロックのマーカーを削除
	diff = cleanCodeBlock(diff)

	return &Proposal{
		Description: strings.TrimSpace(description),
		FilePath:    strings.TrimSpace(filePath),
		Diff:        strings.TrimSpace(diff),
	}, nil
}

// extractSection は、開始マーカーと終了マーカーの間のテキストを抽出します。
func extractSection(text, startMarker, endMarker string) string {
	start := strings.Index(text, startMarker)
	if start == -1 {
		return ""
	}
	start += len(startMarker)

	end := strings.Index(text[start:], endMarker)
	if end == -1 {
		return ""
	}

	return text[start : start+end]
}

// cleanCodeBlock は、コードブロックのマーカー（```）を削除します。
func cleanCodeBlock(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		if idx := strings.Index(text[3:], "```"); idx != -1 {
			// 言語指定（goなど）を含む可能性があるので、最初の改行までをスキップ
			start := strings.Index(text, "\n")
			if start == -1 {
				return ""
			}
			text = text[start+1 : len(text)-3]
		}
	}
	return strings.TrimSpace(text)
}
