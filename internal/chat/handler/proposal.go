package handler

import (
	"fmt"

	"github.com/pankona/makasero/internal/chat/detector"
)

// Approver は、提案の承認を行うインターフェースです。
type Approver interface {
	GetApproval(proposal *detector.Proposal) (bool, error)
}

// Applier は、提案の変更を適用するインターフェースです。
type Applier interface {
	Apply(proposal *detector.Proposal) error
}

// ProposalHandler は、提案の処理を行う構造体です。
type ProposalHandler struct {
	approver Approver
	applier  Applier
}

// NewProposalHandler は、新しいProposalHandlerインスタンスを作成します。
func NewProposalHandler(approver Approver, applier Applier) *ProposalHandler {
	return &ProposalHandler{
		approver: approver,
		applier:  applier,
	}
}

// Handle は、提案の処理を実行します。
// 1. 提案内容の表示
// 2. ユーザーからの承認取得
// 3. 変更の適用
func (h *ProposalHandler) Handle(proposal *detector.Proposal) error {
	// 1. 提案内容の表示
	fmt.Printf("提案内容：\n%s\n", proposal.Description)
	fmt.Printf("対象ファイル：%s\n", proposal.FilePath)
	fmt.Printf("変更内容：\n%s\n", proposal.Diff)

	// 2. ユーザー承認の取得
	approved, err := h.approver.GetApproval(proposal)
	if err != nil {
		return fmt.Errorf("承認取得エラー: %w", err)
	}
	if !approved {
		return nil // 承認されなかった場合は何もせずに終了
	}

	// 3. 変更の適用
	if err := h.applier.Apply(proposal); err != nil {
		return fmt.Errorf("変更適用エラー: %w", err)
	}

	return nil
}
