package detector

import "errors"

// エラー定義
var (
	// ErrInvalidFormat は、提案フォーマットが不正な場合のエラーです。
	ErrInvalidFormat = errors.New("不正な提案フォーマット")
	// ErrFileNotFound は、対象ファイルが見つからない場合のエラーです。
	ErrFileNotFound = errors.New("対象ファイルが見つかりません")
)

// ProposalError は、提案処理中のエラーを表現する構造体です。
type ProposalError struct {
	Phase   string // 検出、表示、適用など
	Message string
	Err     error
}

// Error はエラーメッセージを生成します。
func (e *ProposalError) Error() string {
	return e.Phase + ": " + e.Message + " (" + e.Err.Error() + ")"
}
