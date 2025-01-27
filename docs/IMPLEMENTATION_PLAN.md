# Go CLI実装計画

## 1. コンポーネント構成

### パッケージ構成
```
cmd/roo/          # CLIエントリーポイント
├── main.go       # メインプログラム
└── main_test.go  # メインのテスト

internal/         # 内部パッケージ
├── api/          # APIクライアント
│   ├── client.go
│   └── client_test.go
└── models/       # データモデル
    └── api.go
```

## 2. コアコンポーネント

### 2.1 CLIエントリーポイント（cmd/roo/main.go）
```go
func main() {
    // コマンドライン引数の処理
    command := flag.String("command", "", "Command to execute")
    input := flag.String("input", "", "Input data")
    flag.Parse()

    // APIクライアントの初期化
    client, err := api.NewClient()
    if err != nil {
        // エラー処理
    }

    // コマンドの実行
    result, err := executeCommand(client, *command, *input)
    // 結果の出力
}
```

### 2.2 APIクライアント（internal/api/client.go）
```go
type Client struct {
    httpClient *http.Client
    apiKey     string
    baseURL    string
}

func (c *Client) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
    // OpenAI APIとの通信
    // レスポンスの処理
    // エラーハンドリング
}
```

### 2.3 データモデル（internal/models/api.go）
```go
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

## 3. コマンド実装

### 3.1 explainコマンド
- コードの説明生成
- システムプロンプトの設定
- エラーハンドリング

### 3.2 chatコマンド
- 対話形式の通信
- メッセージ履歴の管理
- JSONパース処理

## 4. 実装手順

### フェーズ1: 基本機能（2-3日）
1. プロジェクト構造の設定
   - ディレクトリ構造
   - go.mod設定
   - Makefile作成

2. APIクライアント実装
   - OpenAI API通信
   - エラーハンドリング
   - タイムアウト設定

3. CLIインターフェース
   - フラグ処理
   - サブコマンド
   - 出力フォーマット

### フェーズ2: 機能拡張（2-3日）
1. コマンド機能
   - explainの実装
   - chatの実装
   - テストの作成

2. エラー処理
   - エラーメッセージ
   - リトライ処理
   - ログ出力

### フェーズ3: 最適化（2-3日）
1. パフォーマンス
   - 並行処理
   - メモリ最適化
   - タイムアウト調整

2. ドキュメント
   - README
   - USAGE.md
   - コードコメント

## 5. テスト計画

### 5.1 ユニットテスト
```go
func TestExecuteCommand(t *testing.T) {
    // コマンド実行のテスト
}

func TestCreateChatCompletion(t *testing.T) {
    // API通信のテスト
}
```

### 5.2 統合テスト
```go
func TestEndToEnd(t *testing.T) {
    // E2Eテストシナリオ
}
```

## 6. 成功基準

1. 基本機能
- すべてのコマンドが正常に動作
- エラー処理が適切に機能
- JSONレスポンスが正しい形式

2. パフォーマンス
- レスポンス時間が適切
- メモリ使用が最適
- エラー発生時の適切な処理

3. 品質
- テストカバレッジ80%以上
- リントエラーなし
- ドキュメント完備

4. 使いやすさ
- 直感的なコマンド
- 分かりやすいエラーメッセージ
- 詳細なヘルプ情報