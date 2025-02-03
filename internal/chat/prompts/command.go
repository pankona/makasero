package prompts

// CommandSystemPrompt は、コマンド推論を行うためのシステムプロンプトです。
const CommandSystemPrompt = `あなたはLinuxコマンドの専門家です。
ユーザーの要求を適切なコマンドに変換してください。

要件：
1. findやgrepなどのLinuxコマンドを使用
2. パイプやリダイレクトなどのシェル機能を活用
3. 読み取り専用の操作のみ許可（rm, mv等の破壊的な操作は禁止）
4. 結果は人間が理解しやすい形式で表示

セキュリティ制約：
1. 実行ディレクトリの制限
   - 指定された作業ディレクトリ内でのみ操作可能
   - 親ディレクトリへのアクセスは禁止
2. コマンドの制限
   - rm, mv, cp等の破壊的な操作は禁止
   - sudoの使用は禁止
   - シェルスクリプトの実行は禁止
3. リソースの制限
   - CPU使用率の制限
   - メモリ使用量の制限
   - 実行時間の制限

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

提案は必ず以下の形式で回答してください：

---COMMAND---
[実行すべきコマンド]
---EXPLANATION---
[このコマンドが何を行い、どのように結果を解釈すべきかの説明]
---VALIDATION---
[コマンドの安全性チェックと制約の確認]
---END---`

// CommandMarkers は、コマンド提案フォーマットで使用されるマーカーを定義する型です。
type CommandMarkers struct {
	Command     string
	Explanation string
	Validation  string
	End         string
}

// DefaultCommandMarkers は、コマンド提案フォーマットで使用される実際のマーカー値を提供します。
var DefaultCommandMarkers = CommandMarkers{
	Command:     "---COMMAND---",
	Explanation: "---EXPLANATION---",
	Validation:  "---VALIDATION---",
	End:         "---END---",
}

// CommandResult はコマンド実行結果を表す構造体です。
type CommandResult struct {
	Success  bool   // 実行が成功したかどうか
	Output   string // 標準出力
	Error    string // 標準エラー出力（エラーが発生した場合）
	ExitCode int    // 終了コード
	Duration int64  // 実行時間（ミリ秒）
}

// ValidationError はコマンドの検証エラーを表す構造体です。
type ValidationError struct {
	Code    string // エラーコード
	Message string // エラーメッセージ
	Details string // 詳細な説明
}

// SecurityConstraints はコマンド実行時のセキュリティ制約を定義します。
type SecurityConstraints struct {
	WorkDir        string   // 作業ディレクトリ
	AllowedCmds    []string // 許可されたコマンド
	BlockedCmds    []string // ブロックされたコマンド
	MaxCPUPercent  int      // 最大CPU使用率（%）
	MaxMemoryMB    int      // 最大メモリ使用量（MB）
	TimeoutSeconds int      // タイムアウト時間（秒）
}

// SystemInfo はシステム情報を表す構造体です。
type SystemInfo struct {
	OS          string            // OSの種類とバージョン
	Shell       string            // シェルの種類
	WorkDir     string            // 作業ディレクトリ
	Environment map[string]string // 環境変数
}
