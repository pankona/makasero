# アーキテクチャ設計

## 全体構造

CLIツールは以下の主要コンポーネントで構成されます：

```
hello-vim-plugin-2/
├── cmd/
│   └── roo/
│       ├── main.go        # CLIのエントリーポイント
│       └── main_test.go   # メインパッケージのテスト
├── internal/
│   ├── api/
│   │   ├── client.go      # OpenAI API クライアント
│   │   └── client_test.go # APIクライアントのテスト
│   └── models/
│       └── api.go         # データモデル定義
├── docs/
│   ├── USAGE.md          # 使用方法ドキュメント
│   └── ...               # その他のドキュメント
└── Makefile              # ビルドスクリプト
```

## コンポーネントの役割

### 1. コア機能

#### cmd/roo/main.go
- CLIのエントリーポイント
- コマンドライン引数の処理
- サブコマンドの実装（explain, chat）
- エラーハンドリング

```go
// コマンドの実行
func executeCommand(client APIClient, command, input string) (interface{}, error) {
    switch command {
    case "explain":
        return executeExplain(client, input)
    case "chat":
        return executeChat(client, input)
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

## データフロー

1. コマンド実行フロー
```
User Input -> CLI引数解析 -> コマンド実行 -> APIリクエスト -> JSON応答
```

2. エラー処理フロー
```
エラー発生 -> エラーラッピング -> JSON形式でエラー返却
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

## テスト戦略

1. ユニットテスト
- 各コンポーネントの独立したテスト
- モックの活用（APIクライアント等）
- テーブル駆動テスト

2. 統合テスト
- コマンド実行の結合テスト
- APIリクエスト/レスポンスのテスト

3. 自動テスト
- GitHub Actionsでの自動テスト
- カバレッジレポート生成

## パフォーマンス考慮事項

- タイムアウト設定の最適化
- 効率的なメモリ使用
- 大きな入力の適切な処理
- レスポンスのストリーミング対応