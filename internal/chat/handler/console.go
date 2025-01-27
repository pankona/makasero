package handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/chat/detector"
)

// ConsoleApprover は、コンソールを通じてユーザーから承認を取得する実装です。
type ConsoleApprover struct {
	tty *os.File
}

// NewConsoleApprover は、新しいConsoleApproverインスタンスを作成します。
func NewConsoleApprover() *ConsoleApprover {
	// /dev/ttyを開いて直接端末からの入力を取得
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		// フォールバックとして標準入力を使用
		tty = os.Stdin
	}
	return &ConsoleApprover{
		tty: tty,
	}
}

// GetApproval は、ユーザーからの承認を取得します。
// y/Y で承認、その他は拒否として扱います。
func (a *ConsoleApprover) GetApproval(proposal *detector.Proposal) (bool, error) {
	fmt.Print("この提案を適用しますか？ [y/N]: ")

	var input string
	fmt.Fscanln(a.tty, &input)

	// 入力を整形して判定
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y", nil
}
