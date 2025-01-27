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
├── models/       # データモデル
│   └── api.go
└── proposal/     # コード提案システム
    ├── manager.go
    ├── diff.go
    └── approval.go
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

### 2.4 コード提案システム（internal/proposal/）
```go
// manager.go
type ProposalManager struct {
    client    *api.Client
    diffUtil  *DiffUtility
    approver  UserApprover
}

// diff.go
type DiffUtility struct {
    // 差分処理機能
}

// approval.go
type UserApprover interface {
    RequestApproval(proposal *CodeProposal) (bool, error)
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

### 3.3 proposeコマンド
- コード提案の生成
- 差分の表示と管理
- ユーザー承認フロー
- コード適用処理

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

### フェーズ3: コード提案システム（3-4日）
1. 基本構造
   - ProposalManagerの実装
   - DiffUtilityの実装
   - UserApproverインターフェース

2. 差分処理
   - パッチ生成
   - 差分表示
   - ファイル操作

3. 承認フロー
   - インタラクティブな承認UI
   - 適用モード選択
   - エラーハンドリング

### フェーズ4: 最適化とテスト（2-3日）
1. パフォーマンス
   - 並行処理
   - メモリ最適化
   - タイムアウト調整

2. テスト拡充
   - 提案システムのユニットテスト
   - 差分処理のテスト
   - 承認フローのテスト

3. ドキュメント
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

func TestProposalManager(t *testing.T) {
    // 提案システムのテスト
}

func TestDiffUtility(t *testing.T) {
    // 差分処理のテスト
}
```

### 5.2 統合テスト
```go
func TestEndToEnd(t *testing.T) {
    // E2Eテストシナリオ
}

func TestProposalWorkflow(t *testing.T) {
    // 提案ワークフローのテスト
}
```

## 6. 成功基準

1. 基本機能
- すべてのコマンドが正常に動作
- エラー処理が適切に機能
- JSONレスポンスが正しい形式

2. コード提案システム
- 正確な差分生成
- 信頼性の高い適用処理
- ユーザーフレンドリーな承認フロー

3. パフォーマンス
- レスポンス時間が適切
- メモリ使用が最適
- エラー発生時の適切な処理

4. 品質
- テストカバレッジ80%以上
- リントエラーなし
- ドキュメント完備

5. 使いやすさ
- 直感的なコマンド
- 分かりやすいエラーメッセージ
- 詳細なヘルプ情報
- 提案内容の明確な表示