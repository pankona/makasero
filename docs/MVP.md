# MVP（Minimum Viable Product）設計書

## 1. 概要

### 1.1 目的
Roo CLIツールの最小機能セットとして、以下を実装します：
- コードの説明機能
- インタラクティブなチャット機能
- JSON形式での入出力

### 1.2 機能範囲
- コマンドラインインターフェース
- OpenAI APIとの通信
- 結果のJSON形式での出力
- エラーハンドリング

## 2. コマンドラインインターフェース

### 2.1 基本コマンド
```bash
# コードの説明
roo -command explain -input "fmt.Println('Hello')"

# チャット機能
roo -command chat -input '[{"role":"user","content":"Hello"}]'
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
└── models/       # データモデル
    └── api.go
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

### 4.3 フェーズ3: 品質向上（1-2日）
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
```

### 5.2 統合テスト
- コマンドライン引数の処理
- APIリクエスト/レスポンス
- エラーケース

## 6. 成功基準

### 6.1 機能要件
- すべてのコマンドが正常に動作
- 適切なJSON形式での出力
- エラーの適切な処理と報告

### 6.2 非機能要件
- レスポンス時間が3秒以内
- メモリ使用が適切
- クロスプラットフォーム対応

### 6.3 品質基準
- テストカバレッジ80%以上
- リントエラーなし
- ドキュメント完備

## 7. 将来の拡張性

### 7.1 追加予定の機能
1. 追加のサブコマンド
2. ストリーミングレスポンス
3. 設定ファイルのサポート
4. 高度なプロンプト管理

### 7.2 設計上の考慮点
- 新しいコマンドの追加が容易
- API抽象化による柔軟性
- 設定のカスタマイズ性