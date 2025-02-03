package prompts

import (
	"fmt"
	"strings"
)

// CommandSystemPrompt は、コマンド推論を行うためのシステムプロンプトです。
const CommandSystemPrompt = `あなたはLinuxコマンドの専門家です。
以下の要求に対して、適切なコマンドを提案してください。

要件：
1. 読み取り専用の操作のみ許可（破壊的な操作は禁止）
2. findやgrepなどの検索系コマンドを使用
3. パイプやリダイレクトは必要に応じて使用
4. 結果は人間が理解しやすい形式で表示

セキュリティ制約：
1. 作業ディレクトリ内でのみ操作
2. 破壊的な操作は禁止
3. 特権コマンドの使用は禁止

コマンド例：
- ファイル検索: find . -name "*.txt" -o -name "*.log"
- サイズ検索: find . -size +100k
- 更新日時: find . -mtime -7
- テキスト検索: grep "検索文字列" ファイル名
- パイプ使用: grep "検索" ファイル | wc -l

システム情報の考慮：
1. OSの種類とバージョン
2. シェルの種類
3. 現在の作業ディレクトリ
4. 利用可能なコマンド

エラーハンドリング：
1. コマンドの構文エラー
2. 権限エラー
3. リソース制限エラー
4. タイムアウトエラー

システム情報：
OS: {{.OS}}
Shell: {{.Shell}}
作業ディレクトリ: {{.WorkDir}}

ユーザーの要求：{{.UserInput}}

以下の形式で回答してください：

---COMMAND---
[実行するコマンド]
---EXPLANATION---
[コマンドの説明]
---VALIDATION---
[安全性の確認]
---END---`

// CommandMarkers はコマンド推論の結果を示すマーカーです。
const (
	CommandMarkerStart       = "---COMMAND---"
	CommandMarkerExplanation = "---EXPLANATION---"
	CommandMarkerValidation  = "---VALIDATION---"
	CommandMarkerEnd         = "---END---"
)

// CommandResult はコマンド推論の結果を表す構造体です。
type CommandResult struct {
	Command     string // 実行するコマンド
	Explanation string // コマンドの説明
	Validation  string // 安全性の確認
}

// ValidationError はバリデーションエラーを表す構造体です。
type ValidationError struct {
	Code    string // エラーコード
	Message string // エラーメッセージ
}

// SecurityConstraints はセキュリティ制約を表す構造体です。
type SecurityConstraints struct {
	AllowedCommands []string          // 許可されたコマンド
	WorkDir         string            // 作業ディレクトリ
	ResourceLimits  map[string]string // リソース制限
}

// SystemInfo はシステム情報を表す構造体です。
type SystemInfo struct {
	OS          string            // OS名
	Shell       string            // シェル名
	WorkDir     string            // 作業ディレクトリ
	Environment map[string]string // 環境変数
}

// CommandPrompt はコマンド推論のためのプロンプトを生成する構造体です。
type CommandPrompt struct {
	userInput string
	sysInfo   SystemInfo
}

// NewCommandPrompt は新しいCommandPromptインスタンスを作成します。
func NewCommandPrompt(userInput string, sysInfo SystemInfo) *CommandPrompt {
	return &CommandPrompt{
		userInput: userInput,
		sysInfo:   sysInfo,
	}
}

// GeneratePrompt はユーザーの入力からプロンプトを生成します。
func (p *CommandPrompt) GeneratePrompt() string {
	var sb strings.Builder

	// システムプロンプトの基本部分
	sb.WriteString("あなたはLinuxコマンドの専門家です。\n")
	sb.WriteString("以下の要求に対して、適切なコマンドを提案してください。\n\n")

	// 要件の記述
	sb.WriteString("要件：\n")
	sb.WriteString("1. 読み取り専用の操作のみ許可（破壊的な操作は禁止）\n")
	sb.WriteString("2. findやgrepなどの検索系コマンドを使用\n")
	sb.WriteString("3. パイプやリダイレクトは必要に応じて使用\n")
	sb.WriteString("4. 結果は人間が理解しやすい形式で表示\n\n")

	// セキュリティ制約
	sb.WriteString("セキュリティ制約：\n")
	sb.WriteString("1. 作業ディレクトリ内でのみ操作\n")
	sb.WriteString("2. 破壊的な操作は禁止\n")
	sb.WriteString("3. 特権コマンドの使用は禁止\n\n")

	// コマンド例の提供
	sb.WriteString("コマンド例：\n")
	sb.WriteString("- ファイル検索: find . -name \"*.txt\" -o -name \"*.log\"\n")
	sb.WriteString("- サイズ検索: find . -size +100k\n")
	sb.WriteString("- 更新日時: find . -mtime -7\n")
	sb.WriteString("- テキスト検索: grep \"検索文字列\" ファイル名\n")
	sb.WriteString("- パイプ使用: grep \"検索\" ファイル | wc -l\n\n")

	// システム情報の考慮
	sb.WriteString("システム情報の考慮：\n")
	sb.WriteString("1. OSの種類とバージョン\n")
	sb.WriteString("2. シェルの種類\n")
	sb.WriteString("3. 現在の作業ディレクトリ\n")
	sb.WriteString("4. 利用可能なコマンド\n\n")

	// エラーハンドリング
	sb.WriteString("エラーハンドリング：\n")
	sb.WriteString("1. コマンドの構文エラー\n")
	sb.WriteString("2. 権限エラー\n")
	sb.WriteString("3. リソース制限エラー\n")
	sb.WriteString("4. タイムアウトエラー\n\n")

	// システム情報
	sb.WriteString(fmt.Sprintf("システム情報：\n"))
	sb.WriteString(fmt.Sprintf("OS: %s\n", p.sysInfo.OS))
	sb.WriteString(fmt.Sprintf("Shell: %s\n", p.sysInfo.Shell))
	sb.WriteString(fmt.Sprintf("作業ディレクトリ: %s\n\n", p.sysInfo.WorkDir))

	// ユーザーの要求
	sb.WriteString(fmt.Sprintf("ユーザーの要求：%s\n\n", p.userInput))

	// 回答フォーマット
	sb.WriteString("以下の形式で回答してください：\n\n")
	sb.WriteString("---COMMAND---\n")
	sb.WriteString("[実行するコマンド]\n")
	sb.WriteString("---EXPLANATION---\n")
	sb.WriteString("[コマンドの説明]\n")
	sb.WriteString("---VALIDATION---\n")
	sb.WriteString("[安全性の確認]\n")
	sb.WriteString("---END---\n")

	return sb.String()
}
