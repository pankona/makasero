# MVP（Minimum Viable Product）設計書

## 1. 概要

### 1.1 目的
Roo CLIツールの最小機能セットとして、以下を実装します：
- コードの説明機能
- インタラクティブなチャット機能
- コード提案と適用機能
- JSON形式での入出力

### 1.2 機能範囲
- コマンドラインインターフェース
- OpenAI APIとの通信
- 結果のJSON形式での出力
- エラーハンドリング
- コード提案の生成と適用
- ユーザー承認フロー

## 2. コマンドラインインターフェース

### 2.1 基本コマンド
```bash
# コードの説明
roo -command explain -input "fmt.Println('Hello')"

# チャット機能
roo -command chat -input '[{"role":"user","content":"Hello"}]'

# コード提案
roo -command propose -input "path/to/file.go" -mode "patch"  # または "full"
```

### 2.2 出力フォーマット
```json
{
  "success": true,
  "data": "応答内容",
  "error": null
}
```

### 2.3 エラー出力
```json
{
  "success": false,
  "data": null,
  "error": "エラーメッセージ"
}
```

### 2.4 提案フォーマット
```json
{
  "success": true,
  "data": {
    "original": "元のコード",
    "proposed": "提案されたコード",
    "diff": "差分内容",
    "description": "変更の説明"
  }
}
```

## 3. 技術実装

### 3.1 プロジェクト構造
```
cmd/roo/
├── main.go       # エントリーポイント
└── main_test.go  # メインのテスト

internal/
├── api/          # API通信
│   ├── client.go
│   └── client_test.go
├── models/       # データモデル
│   └── api.go
└── proposal/     # コード提案システム
    ├── manager.go
    ├── diff.go
    └── approval.go
```

### 3.2 データモデル
```go
// チャットメッセージ
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// APIレスポンス
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// チャットリクエスト
type ChatRequest struct {
    Model    string        `json:"model"`
    Messages []ChatMessage `json:"messages"`
}

// コード提案
type CodeProposal struct {
    OriginalCode string `json:"original_code"`
    ProposedCode string `json:"proposed_code"`
    FilePath     string `json:"file_path"`
    DiffContent  string `json:"diff_content"`
    ApplyMode    string `json:"apply_mode"` // "patch" or "full"
}
```

## 4. 実装フロー

### 4.1 フェーズ1: 基本構造（2-3日）
1. プロジェクト設定
   - ディレクトリ構造
   - 依存関係管理
   - ビルドスクリプト

2. APIクライアント
   - OpenAI API通信
   - エラーハンドリング
   - タイムアウト設定

### 4.2 フェーズ2: コア機能（2-3日）
1. コマンド実装
   - explainコマンド
   - chatコマンド
   - 引数処理

2. 出力処理
   - JSON形式の出力
   - エラー出力
   - ログ出力

### 4.3 フェーズ3: 提案システム（2-3日）
1. 基本機能
   - コード分析
   - 提案生成
   - 差分処理

2. 承認フロー
   - ユーザー確認
   - 変更適用
   - ロールバック機能

### 4.4 フェーズ4: 品質向上（1-2日）
1. テスト実装
   - ユニットテスト
   - 統合テスト
   - カバレッジ改善

2. ドキュメント
   - README
   - 使用例
   - エラーメッセージ

## 5. テスト計画

### 5.1 ユニットテスト
```go
func TestExecuteCommand(t *testing.T) {
    // コマンド実行のテスト
    tests := []struct {
        name    string
        command string
        input   string
        want    interface{}
        wantErr bool
    }{
        // テストケース
    }
    // ...
}

func TestCreateChatCompletion(t *testing.T) {
    // API通信のテスト
}

func TestCodeProposal(t *testing.T) {
    // 提案システムのテスト
}
```

### 5.2 統合テスト
- コマンドライン引数の処理
- APIリクエスト/レスポンス
- エラーケース
- 提案フローの検証

## 6. 成功基準

### 6.1 機能要件
- すべてのコマンドが正常に動作
- 適切なJSON形式での出力
- エラーの適切な処理と報告
- コード提案の正確性
- 承認フローの信頼性

### 6.2 非機能要件
- レスポンス時間が3秒以内
- メモリ使用が適切
- クロスプラットフォーム対応
- 差分適用の安全性

### 6.3 品質基準
- テストカバレッジ80%以上
- リントエラーなし
- ドキュメント完備
- コード変更の追跡可能性

## 7. 将来の拡張性

### 7.1 追加予定の機能
1. 追加のサブコマンド
2. ストリーミングレスポンス
3. 設定ファイルのサポート
4. 高度なプロンプト管理
5. バッチ処理による複数ファイルの提案
6. カスタム提案テンプレート
7. 変更履歴の管理

### 7.2 設計上の考慮点
- 新しいコマンドの追加が容易
- API抽象化による柔軟性
- 設定のカスタマイズ性
- 提案エンジンの拡張性
- 差分アルゴリズムの交換可能性