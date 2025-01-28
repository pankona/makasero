# アーキテクチャ設計

## 全体構造

CLIツールは以下の主要コンポーネントで構成されます：

```
makasero/
├── cmd/
│   └── makasero/
│       ├── main.go        # CLIのエントリーポイント
│       └── main_test.go   # メインパッケージのテスト
├── internal/
│   ├── api/
│   │   ├── client.go      # OpenAI API クライアント
│   │   └── client_test.go # APIクライアントのテスト
│   ├── models/
│   │   └── api.go         # データモデル定義
│   └── proposal/          # 新規: コード提案管理
│       ├── manager.go     # コード提案マネージャー
│       └── diff.go        # 差分処理ユーティリティ
├── docs/
│   ├── USAGE.md          # 使用方法ドキュメント
│   └── ...               # その他のドキュメント
└── Makefile              # ビルドスクリプト
```

## コンポーネントの役割

### 1. コア機能

#### cmd/makasero/main.go
- CLIのエントリーポイント
- コマンドライン引数の処理
- サブコマンドの実装（explain, chat, propose）
- エラーハンドリング

```go
// コマンドの実行
func executeCommand(client APIClient, command, input string) (interface{}, error) {
    switch command {
    case "explain":
        return executeExplain(client, input)
    case "chat":
        return executeChat(client, input)
    case "propose":
        return executePropose(client, input)
    default:
        return nil, fmt.Errorf("unknown command: %s", command)
    }
}
```

#### internal/api/client.go
- OpenAI APIとの通信処理
- リクエスト/レスポンスの処理
- エラーハンドリング
- タイムアウト管理

```go
// APIクライアント
type Client struct {
    httpClient *http.Client
    apiKey     string
    baseURL    string
}

// チャット完了リクエスト
func (c *Client) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
    // APIリクエストの処理
}
```

#### internal/models/api.go
- データ構造の定義
- JSONシリアライズ/デシリアライズ
- バリデーション

```go
// チャットメッセージモデル
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// レスポンス構造
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### 2. コード提案システム

#### internal/proposal/manager.go
- コード提案の管理
- 提案の検証と承認フロー
- 差分の生成と適用

```go
// コード提案マネージャー
type ProposalManager struct {
    client    APIClient
    diffUtil  *DiffUtility
    approver  UserApprover
}

// 提案構造体
type CodeProposal struct {
    OriginalCode string
    ProposedCode string
    FilePath     string
    DiffContent  string
    ApplyMode    ApplyMode // Patch or FullRewrite
}

// 承認インターフェース
type UserApprover interface {
    RequestApproval(proposal *CodeProposal) (bool, error)
}
```

#### internal/proposal/diff.go
- 差分の生成と解析
- パッチの適用
- ファイル操作

```go
// 差分ユーティリティ
type DiffUtility struct {
    // 差分生成と適用のための機能
}

// 適用モード
type ApplyMode int

const (
    ApplyModePatch ApplyMode = iota
    ApplyModeFullRewrite
)
```

## データフロー

1. コード提案フロー
```
ユーザー入力 -> AI提案生成 -> 差分生成 -> ユーザー承認 -> コード適用
```

2. 承認フロー
```
提案表示 -> 差分プレビュー -> ユーザー確認 -> 適用実行
```

## エラー処理

- 各レイヤーでの適切なエラーハンドリング
- エラーのラッピングとコンテキスト追加
- ユーザーフレンドリーなエラーメッセージ
- 詳細なエラーログ

## 設定システム

```go
const (
    defaultTimeout = 30 * time.Second
    defaultBaseURL = "https://api.openai.com/v1"
)

// 環境変数による設定
// OPENAI_API_KEY: API認証キー
```

## 拡張性

- 新しいサブコマンドの追加
- 異なるAIモデルのサポート
- カスタムプロンプトの実装
- 出力フォーマットの拡張
- コード提案の適用方法のカスタマイズ

## テスト戦略

1. ユニットテスト
- 各コンポーネントの独立したテスト
- モックの活用（APIクライアント等）
- テーブル駆動テスト
- 提案システムのテスト

2. 統合テスト
- コマンド実行の結合テスト
- APIリクエスト/レスポンスのテスト
- 差分適用のテスト

3. 自動テスト
- GitHub Actionsでの自動テスト
- カバレッジレポート生成

## パフォーマンス考慮事項

- タイムアウト設定の最適化
- 効率的なメモリ使用
- 大きな入力の適切な処理
- レスポンスのストリーミング対応
- 差分生成の効率化