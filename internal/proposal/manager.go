package proposal

import (
	"fmt"
	"io/ioutil"

	"github.com/pankona/makasero/internal/api"
	"github.com/pankona/makasero/internal/models"
)

// ApplyMode は提案の適用モードを表します
type ApplyMode string

const (
	// ApplyModePatch はパッチモードでの適用を表します
	ApplyModePatch ApplyMode = "patch"
	// ApplyModeFull は全体書き換えモードでの適用を表します
	ApplyModeFull ApplyMode = "full"
)

// CodeProposal はコード提案を表す構造体です
type CodeProposal struct {
	OriginalCode string    // 元のコード
	ProposedCode string    // 提案されたコード
	FilePath     string    // 対象ファイルのパス
	DiffContent  string    // 差分内容
	ApplyMode    ApplyMode // 適用モード
	Description  string    // 変更の説明
}

// UserApprover はユーザーの承認を得るためのインターフェースです
type UserApprover interface {
	RequestApproval(proposal *CodeProposal) (bool, error)
}

// Manager はコード提案を管理するための構造体です
type Manager struct {
	client   api.APIClient
	diffUtil *DiffUtility
	approver UserApprover
}

// NewManager は新しいManagerインスタンスを作成します
func NewManager(client api.APIClient, approver UserApprover) *Manager {
	return &Manager{
		client:   client,
		diffUtil: NewDiffUtility(),
		approver: approver,
	}
}

// GenerateProposal は指定されたファイルに対するコード提案を生成します
func (m *Manager) GenerateProposal(filePath string, mode ApplyMode) (*CodeProposal, error) {
	// ファイルの読み込み
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルの読み込みに失敗: %w", err)
	}

	// AIにコード改善の提案を要求
	proposedCode, description, err := m.requestAIProposal(string(content))
	if err != nil {
		return nil, fmt.Errorf("AI提案の生成に失敗: %w", err)
	}

	// 差分の生成
	diff, err := m.diffUtil.GenerateDiff(string(content), proposedCode)
	if err != nil {
		return nil, fmt.Errorf("差分の生成に失敗: %w", err)
	}

	return &CodeProposal{
		OriginalCode: string(content),
		ProposedCode: proposedCode,
		FilePath:     filePath,
		DiffContent:  diff,
		ApplyMode:    mode,
		Description:  description,
	}, nil
}

// ApplyProposal は提案された変更を適用します
func (m *Manager) ApplyProposal(proposal *CodeProposal) error {
	// ユーザーの承認を得る
	approved, err := m.approver.RequestApproval(proposal)
	if err != nil {
		return fmt.Errorf("承認プロセスでエラー: %w", err)
	}

	if !approved {
		return fmt.Errorf("ユーザーが変更を承認しませんでした")
	}

	// 変更の適用
	var newContent string
	if proposal.ApplyMode == ApplyModePatch {
		newContent, err = m.diffUtil.ApplyPatch(proposal.OriginalCode, proposal.DiffContent)
		if err != nil {
			return fmt.Errorf("パッチの適用に失敗: %w", err)
		}
	} else {
		newContent = proposal.ProposedCode
	}

	// ファイルの書き込み
	err = ioutil.WriteFile(proposal.FilePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗: %w", err)
	}

	return nil
}

// requestAIProposal はAIにコード改善の提案を要求します
func (m *Manager) requestAIProposal(code string) (string, string, error) {
	messages := []models.ChatMessage{
		{
			Role:    "system",
			Content: "あなたはコードレビューと改善提案を行う専門家です。",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("以下のコードを改善してください:\n\n%s", code),
		},
	}

	response, err := m.client.CreateChatCompletion(messages)
	if err != nil {
		return "", "", err
	}

	// TODO: レスポンスのパースと提案の抽出を実装
	// 現在は単純な実装
	return response, "コードの改善提案", nil
}
