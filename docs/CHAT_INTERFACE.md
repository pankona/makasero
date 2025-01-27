# チャットインターフェースと提案システム

## 概要

チャットインターフェースを通じてコード提案を統合的に扱う新しいアーキテクチャを提案します。
AIプロンプトを活用して、明確な構造を持つレスポンスを得ることで、信頼性の高い提案システムを実現します。

## プロンプト設計

### 1. システムプロンプト

```go
const systemPrompt = `あなたはコードレビューと改善提案を行う専門家です。
コードの改善提案を行う場合は、必ず以下の形式で回答してください：

---PROPOSAL---
[提案の説明]

---FILE---
[対象ファイルパス]

---DIFF---
[変更内容]
---END---

提案ではない場合は、通常の形式で回答してください。`
```

### 2. レスポンス構造

提案を含むレスポンスの例：
```
コードを分析しました。以下の改善を提案します：

---PROPOSAL---
エラーハンドリングを改善し、ログ出力を追加します。

---FILE---
internal/api/client.go

---DIFF---
@@ -10,6 +10,7 @@
 func (c *Client) Execute() error {
-    result := process()
+    result, err := process()
+    if err != nil {
+        log.Printf("処理エラー: %v", err)
+        return fmt.Errorf("実行エラー: %w", err)
+    }
     return nil
 }
---END---
```

通常のレスポンス例：
```
はい、その実装は正しいです。特に改善の必要はありません。
```

## 実装設計

### 1. 提案検出

```go
// internal/chat/proposal.go
type ProposalDetector struct {
    // 提案のマーカー
    proposalMarker    string
    fileMarker       string
    diffMarker       string
    endMarker        string
}

func (d *ProposalDetector) IsProposal(response string) bool {
    return strings.Contains(response, "---PROPOSAL---")
}

func (d *ProposalDetector) ExtractProposal(response string) (*CodeProposal, error) {
    if !d.IsProposal(response) {
        return nil, nil
    }

    // マーカーに基づいて各セクションを抽出
    proposal := &CodeProposal{
        Description: extractSection(response, "PROPOSAL", "FILE"),
        FilePath:    extractSection(response, "FILE", "DIFF"),
        Diff:        extractSection(response, "DIFF", "END"),
    }

    return proposal, nil
}
```

### 2. チャット実行

```go
// internal/chat/executor.go
func executeChat(client APIClient, input string) (string, error) {
    // システムプロンプトを追加
    messages := []ChatMessage{
        {
            Role:    "system",
            Content: systemPrompt,
        },
        {
            Role:    "user",
            Content: input,
        },
    }

    // AIとの対話
    response, err := client.CreateChatCompletion(messages)
    if err != nil {
        return "", err
    }

    // 提案の検出と処理
    detector := NewProposalDetector()
    if proposal, err := detector.ExtractProposal(response); err != nil {
        return "", err
    } else if proposal != nil {
        if err := handleProposal(proposal); err != nil {
            return "", err
        }
    }

    return response, nil
}
```

### 3. 提案処理

```go
// internal/chat/handler.go
func handleProposal(proposal *CodeProposal) error {
    // 1. 提案内容の表示
    fmt.Printf("提案内容：\n%s\n", proposal.Description)
    fmt.Printf("対象ファイル：%s\n", proposal.FilePath)
    fmt.Printf("変更内容：\n%s\n", proposal.Diff)

    // 2. ユーザー承認の取得
    if !getUserApproval() {
        return nil
    }

    // 3. 変更の適用
    return applyChanges(proposal)
}
```

## エラーハンドリング

```go
// internal/chat/errors.go
type ProposalError struct {
    Phase   string // 検出、表示、適用
    Message string
    Err     error
}

var (
    ErrInvalidFormat    = errors.New("不正な提案フォーマット")
    ErrFileNotFound     = errors.New("対象ファイルが見つかりません")
    ErrApplyFailed      = errors.New("変更の適用に失敗しました")
)
```

## テスト戦略

### 1. プロンプトテスト

```go
func TestProposalDetection(t *testing.T) {
    cases := []struct {
        name     string
        response string
        want     bool
    }{
        {
            name:     "提案を含むレスポンス",
            response: "---PROPOSAL---\n説明\n---FILE---\npath\n---DIFF---\n差分\n---END---",
            want:     true,
        },
        {
            name:     "通常のレスポンス",
            response: "はい、その実装で問題ありません。",
            want:     false,
        },
    }
    // テスト実装
}
```

### 2. 統合テスト

```go
func TestChatWithProposal(t *testing.T) {
    // 1. モックAIクライアントの設定
    // 2. チャット実行
    // 3. 提案検出の確認
    // 4. 変更適用の確認
}
```

## 利点

1. 信頼性
- 明確な構造による確実な提案検出
- パース処理の簡素化
- エラーの少ない実装

2. 保守性
- シンプルなコード構造
- 理解しやすい処理フロー
- テストの容易さ

3. 拡張性
- 新しい提案形式の追加が容易
- プロンプトの柔軟な調整
- 処理フローのカスタマイズ

## 制限事項

1. プロンプトの依存性
- AIモデルの応答品質に依存
- プロンプトの継続的な改善が必要

2. エラー処理
- 不完全なレスポンス形式への対応
- ファイル操作の失敗への対応

## 今後の展開

1. プロンプトの最適化
- より正確な提案生成
- コンテキストの活用
- エラー検出の改善

2. ユーザーインターフェース
- リッチな差分表示
- インタラクティブな編集
- 提案履歴の管理