package proposal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	client    api.APIClient
	diffUtil  *DiffUtility
	approver  UserApprover
	backupDir string
}

// NewManager は新しいManagerインスタンスを作成します
func NewManager(client api.APIClient, approver UserApprover, backupDir string) *Manager {
	return &Manager{
		client:    client,
		diffUtil:  NewDiffUtility(),
		approver:  approver,
		backupDir: backupDir,
	}
}

// createBackup はファイルのバックアップを作成します
func (m *Manager) createBackup(filePath string) (string, error) {
	if m.backupDir == "" {
		return "", nil // バックアップディレクトリが指定されていない場合はスキップ
	}

	// バックアップディレクトリの作成
	err := os.MkdirAll(m.backupDir, 0755)
	if err != nil {
		return "", fmt.Errorf("バックアップディレクトリの作成に失敗: %w", err)
	}

	// バックアップファイル名の生成
	fileName := filepath.Base(filePath)
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(m.backupDir, fmt.Sprintf("%s.%s.bak", fileName, timestamp))

	// ファイルのコピー
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("バックアップ元ファイルの読み込みに失敗: %w", err)
	}

	err = ioutil.WriteFile(backupPath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("バックアップファイルの作成に失敗: %w", err)
	}

	return backupPath, nil
}

// restoreBackup はバックアップからファイルを復元します
func (m *Manager) restoreBackup(backupPath, originalPath string) error {
	if backupPath == "" {
		return nil // バックアップが存在しない場合はスキップ
	}

	content, err := ioutil.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("バックアップファイルの読み込みに失敗: %w", err)
	}

	err = ioutil.WriteFile(originalPath, content, 0644)
	if err != nil {
		return fmt.Errorf("ファイルの復元に失敗: %w", err)
	}

	return nil
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
	// バックアップの作成
	backupPath, err := m.createBackup(proposal.FilePath)
	if err != nil {
		return fmt.Errorf("バックアップの作成に失敗: %w", err)
	}

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
			// エラーが発生した場合、バックアップから復元
			if backupPath != "" {
				if restoreErr := m.restoreBackup(backupPath, proposal.FilePath); restoreErr != nil {
					return fmt.Errorf("パッチの適用に失敗し、バックアップからの復元にも失敗: %v, %w", restoreErr, err)
				}
			}
			return fmt.Errorf("パッチの適用に失敗: %w", err)
		}
	} else {
		newContent = proposal.ProposedCode
	}

	// ファイルの書き込み
	err = ioutil.WriteFile(proposal.FilePath, []byte(newContent), 0644)
	if err != nil {
		// エラーが発生した場合、バックアップから復元
		if backupPath != "" {
			if restoreErr := m.restoreBackup(backupPath, proposal.FilePath); restoreErr != nil {
				return fmt.Errorf("ファイルの書き込みに失敗し、バックアップからの復元にも失敗: %v, %w", restoreErr, err)
			}
		}
		return fmt.Errorf("ファイルの書き込みに失敗: %w", err)
	}

	return nil
}

// requestAIProposal はAIにコード改善の提案を要求します
func (m *Manager) requestAIProposal(code string) (string, string, error) {
	messages := []models.ChatMessage{
		{
			Role: "system",
			Content: `あなたはコードレビューと改善提案を行う専門家です。
コードの改善提案を行う際は、以下の点に注目してください：
1. コードの品質と可読性
2. パフォーマンスの最適化
3. エラーハンドリング
4. ベストプラクティスの適用

提案は以下の形式で行ってください：

---PROPOSAL---
[提案の説明]

---CODE---
[改善後のコード]

---END---`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("以下のコードを改善してください:\n\n```go\n%s\n```", code),
		},
	}

	response, err := m.client.CreateChatCompletion(messages)
	if err != nil {
		return "", "", fmt.Errorf("AI提案の生成に失敗: %w", err)
	}

	// レスポンスのパース
	proposedCode, description, err := parseAIResponse(response)
	if err != nil {
		return "", "", fmt.Errorf("AI応答のパースに失敗: %w", err)
	}

	return proposedCode, description, nil
}

// parseAIResponse はAIの応答から提案コードと説明を抽出します
func parseAIResponse(response string) (string, string, error) {
	// プロポーザルセクションの抽出
	proposalStart := strings.Index(response, "---PROPOSAL---")
	codeStart := strings.Index(response, "---CODE---")
	endMarker := strings.Index(response, "---END---")

	if proposalStart == -1 || codeStart == -1 || endMarker == -1 {
		return "", "", fmt.Errorf("不正なレスポンス形式")
	}

	// 説明の抽出
	description := strings.TrimSpace(response[proposalStart+len("---PROPOSAL---") : codeStart])

	// コードの抽出
	code := strings.TrimSpace(response[codeStart+len("---CODE---") : endMarker])

	return code, description, nil
}
