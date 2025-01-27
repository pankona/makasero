package proposal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CLIApprover はコマンドラインベースの承認フローを実装します
type CLIApprover struct {
	reader *bufio.Reader
}

// NewCLIApprover は新しいCLIApproverインスタンスを作成します
func NewCLIApprover() *CLIApprover {
	return &CLIApprover{
		reader: bufio.NewReader(os.Stdin),
	}
}

// RequestApproval はユーザーに変更の承認を求めます
func (a *CLIApprover) RequestApproval(proposal *CodeProposal) (bool, error) {
	// 提案内容の表示
	fmt.Printf("\n=== コード提案 ===\n\n")
	fmt.Printf("対象ファイル: %s\n", proposal.FilePath)
	fmt.Printf("適用モード: %s\n", proposal.ApplyMode)
	fmt.Printf("\n説明:\n%s\n", proposal.Description)

	// 差分の表示
	diffUtil := NewDiffUtility()
	fmt.Printf("\n%s\n", diffUtil.FormatDiff(proposal.OriginalCode, proposal.ProposedCode))

	// ユーザーの承認を求める
	fmt.Print("\nこの変更を適用しますか？ [y/N]: ")
	response, err := a.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("ユーザー入力の読み取りに失敗: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// ConsoleUserApprover はCLIApproverのインスタンスを返します
func ConsoleUserApprover() UserApprover {
	return NewCLIApprover()
}
