# 実装計画：プロンプトベースの提案システム

## 1. コアコンポーネント

### プロンプト設計
```go
// internal/chat/prompts/proposal.go
const (
    ProposalSystemPrompt = `あなたはコードレビューと改善提案を行う専門家です。
コードの改善提案を行う場合は、必ず以下の形式で回答してください：

---PROPOSAL---
[提案の説明]

---FILE---
[対象ファイルパス]

---DIFF---
[変更内容]
---END---

提案ではない場合は、通常の形式で回答してください。`
)
```

### 提案検出
```go
// internal/chat/detector/proposal.go
type ProposalDetector struct {
    markers struct {
        proposal string
        file     string
        diff     string
        end      string
    }
}

func (d *ProposalDetector) IsProposal(response string) bool {
    return strings.Contains(response, d.markers.proposal)
}

func (d *ProposalDetector) Extract(response string) (*Proposal, error) {
    if !d.IsProposal(response) {
        return nil, nil
    }
    // マーカーに基づいて提案を抽出
    return extractProposal(response)
}
```

### 提案処理
```go
// internal/chat/handler/proposal.go
type ProposalHandler struct {
    approver Approver
    applier  Applier
}

func (h *ProposalHandler) Handle(proposal *Proposal) error {
    // 1. 提案の表示
    // 2. 承認の取得
    // 3. 変更の適用
}
```

## 2. インターフェース定義

### チャットクライアント
```go
// internal/chat/client.go
type ChatClient interface {
    Execute(prompt string, messages []Message) (string, error)
}

type Message struct {
    Role    string
    Content string
}
```

### 承認インターフェース
```go
// internal/chat/approval.go
type Approver interface {
    GetApproval(proposal *Proposal) (bool, error)
}

type ConsoleApprover struct{}
```

### 変更適用インターフェース
```go
// internal/chat/apply.go
type Applier interface {
    Apply(proposal *Proposal) error
}

type FileApplier struct{}
```

## 3. メインフロー実装

### チャット実行
```go
// cmd/roo/main.go
func executeChat(client ChatClient, input string) error {
    // 1. システムプロンプトの設定
    messages := []Message{
        {Role: "system", Content: ProposalSystemPrompt},
        {Role: "user", Content: input},
    }

    // 2. チャットの実行
    response, err := client.Execute(input, messages)
    if err != nil {
        return err
    }

    // 3. 提案の検出と処理
    detector := NewProposalDetector()
    if proposal, err := detector.Extract(response); err != nil {
        return err
    } else if proposal != nil {
        handler := NewProposalHandler()
        return handler.Handle(proposal)
    }

    return nil
}
```

## 4. エラー処理

### カスタムエラー
```go
// internal/chat/errors.go
type ProposalError struct {
    Phase   string
    Message string
    Err     error
}

var (
    ErrInvalidFormat = errors.New("不正な提案フォーマット")
    ErrFileNotFound  = errors.New("対象ファイルが見つかりません")
    ErrApplyFailed   = errors.New("変更の適用に失敗しました")
)
```

## 5. テスト実装

### プロンプトテスト
```go
// internal/chat/prompts/proposal_test.go
func TestProposalFormat(t *testing.T) {
    // システムプロンプトのテスト
    // レスポンスフォーマットの検証
}
```

### 検出テスト
```go
// internal/chat/detector/proposal_test.go
func TestProposalDetection(t *testing.T) {
    // 提案検出のテスト
    // 各セクションの抽出テスト
}
```

### 統合テスト
```go
// cmd/roo/main_test.go
func TestChatWithProposal(t *testing.T) {
    // エンドツーエンドテスト
    // モックレスポンスの使用
}
```

## 6. 実装手順

1. 基盤実装（2日）
   - プロンプト設計の実装
   - 基本インターフェースの定義
   - エラー型の定義

2. コア機能実装（3日）
   - 提案検出の実装
   - 提案処理の実装
   - ファイル操作の実装

3. テスト実装（2日）
   - ユニットテストの作成
   - 統合テストの作成
   - エッジケースの検証

4. リファクタリング（1日）
   - コードの最適化
   - エラー処理の改善
   - パフォーマンスの調整

5. ドキュメント更新（1日）
   - コメントの追加
   - README の更新
   - 使用例の追加

## 7. 成功基準

### 機能要件
- [ ] 提案の正確な検出
- [ ] 安全なファイル操作
- [ ] 明確なエラーメッセージ

### 非機能要件
- [ ] 90%以上のテストカバレッジ
- [ ] エラー時の適切なフォールバック
- [ ] 分かりやすいドキュメント

## 8. 将来の拡張性

1. プロンプトの拡張
- 複数ファイルの提案対応
- コンテキストの活用
- カスタムフォーマット

2. 機能の拡張
- 提案履歴の管理
- バッチ処理対応
- IDE統合

3. UI/UX改善
- リッチな差分表示
- インタラクティブな編集
- 進捗表示