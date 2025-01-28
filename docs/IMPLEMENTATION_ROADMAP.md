# 実装ロードマップ

## フェーズ1: 基盤整備

### 1. パッケージの再構成
- `internal/proposal` パッケージの機能を `internal/chat` に統合
- 提案関連の型とインターフェースの移行
- 既存の提案処理ロジックのリファクタリング

### 2. 新しいインターフェースの実装
```go
// internal/chat/proposal.go
type ProposalDetector interface {
    DetectProposal(response string) (*CodeProposal, error)
}

type ProposalHandler interface {
    HandleProposal(proposal *CodeProposal) error
}
```

## フェーズ2: チャット機能の拡張

### 1. executeChat の改善
```go
// cmd/makasero/main.go の変更
func executeChat(client APIClient, input string) (string, error) {
    response, err := client.CreateChatCompletion(messages)
    if err != nil {
        return "", err
    }

    // 提案の検出と処理を統合
    if err := handleChatResponse(response); err != nil {
        return "", err
    }

    return response, nil
}
```

### 2. 提案検出ロジックの実装
```go
// internal/chat/detector.go
func detectProposal(response string) (*CodeProposal, error) {
    // Markdown パーサーの実装
    // コードブロックの検出
    // 差分の解析
    return proposal, nil
}
```

## フェーズ3: 提案処理の改善

### 1. 差分処理の最適化
```go
// internal/chat/diff.go
func generateDiff(original, proposed string) (string, error) {
    // より効率的な差分生成
    // 可読性の高い形式
    return diff, nil
}
```

### 2. ファイル操作の安全性向上
```go
// internal/chat/file.go
func applyChanges(proposal *CodeProposal) error {
    // バックアップの作成
    // 安全な書き込み
    // エラー時のロールバック
    return nil
}
```

## フェーズ4: CLI インターフェースの更新

### 1. コマンドラインオプションの変更
```go
// cmd/makasero/main.go
func main() {
    // propose コマンドの削除
    // chat コマンドのオプション追加
    // 新しいフラグの追加
}
```

### 2. ヘルプメッセージの更新
```go
const helpText = `
Usage: makasero [command] [options]

Commands:
  chat     AIとチャットし、コード改善提案を受け取る
  explain  コードの説明を取得する

Options:
  --input    入力テキストまたはファイル
  --format   出力フォーマット (text/json)
`
```

## フェーズ5: テストの更新

### 1. 新しいテストケースの追加
```go
// internal/chat/detector_test.go
func TestProposalDetection(t *testing.T) {
    // 様々なフォーマットのテスト
    // エッジケースの処理
    // エラー条件のテスト
}
```

### 2. 統合テストの更新
```go
// cmd/makasero/main_test.go
func TestChatWithProposal(t *testing.T) {
    // エンドツーエンドのテスト
    // 実際のファイル操作を含むテスト
}
```

## フェーズ6: ドキュメントの更新

### 1. ユーザードキュメント
- README.mdの更新
- 新しい使用方法の説明
- 例示の追加

### 2. 開発者ドキュメント
- アーキテクチャの説明
- 拡張方法の説明
- APIリファレンス

## 実装スケジュール

1. フェーズ1: 1週間
   - パッケージ構造の変更
   - 基本インターフェースの実装

2. フェーズ2: 1週間
   - チャット機能の拡張
   - 提案検出の実装

3. フェーズ3: 1週間
   - 差分処理の改善
   - ファイル操作の実装

4. フェーズ4: 3日
   - CLIの更新
   - ユーザーインターフェースの調整

5. フェーズ5: 3日
   - テストの作成と更新
   - バグ修正

6. フェーズ6: 2日
   - ドキュメントの更新
   - 最終調整

## リスクと対策

### 1. 後方互換性
- 既存のスクリプトやツールへの影響
- 移行期間の設定
- 互換モードの提供

### 2. パフォーマンス
- 大きなファイルの処理
- メモリ使用量の最適化
- 応答時間の維持

### 3. エラー処理
- 予期せぬフォーマット
- ファイル操作の失敗
- ネットワークの問題

## 成功基準

1. 機能面
- すべての既存機能が新しいフローで動作
- エラー処理の完全性
- テストカバレッジ80%以上

2. 性能面
- レスポンス時間の維持
- メモリ使用量の制御
- 安定した動作

3. ユーザビリティ
- 直感的な使用感
- 明確なエラーメッセージ
- 充実したドキュメント