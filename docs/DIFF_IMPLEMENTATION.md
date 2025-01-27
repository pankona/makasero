# 差分適用戦略の改善

## 現状の課題
現在の実装（internal/proposal/diff.go）は基本的な差分処理を`go-diff/diffmatchpatch`を使用して行っていますが、以下の点で改善の余地があります：

1. 正確な検索と置換の保証が不十分
2. インデントの保持が明示的に処理されていない
3. エラー報告が最小限

## 改善案

Roo-Code-mainの実装を参考に、以下の機能を追加することを提案します：

### 1. SearchReplaceDiffStrategy の導入

```go
// SearchReplaceDiffStrategy は検索と置換に基づく差分適用戦略を提供します
type SearchReplaceDiffStrategy struct {
    fuzzyThreshold float64
    bufferLines    int
}

// ApplyDiff は与えられた差分を適用します
func (s *SearchReplaceDiffStrategy) ApplyDiff(original, searchContent, replaceContent string, startLine, endLine int) (string, error) {
    // 1. 類似度チェック
    // 2. インデント保持
    // 3. 行末文字の保持
    // 4. エラー時の詳細な診断情報
}
```

### 2. 主な改善点

#### 2.1 類似度チェック
- Levenshtein距離を使用して検索対象との類似度を計算
- 設定可能な閾値（fuzzyThreshold）による柔軟な一致判定
- 部分一致の場合は詳細な診断情報を提供

#### 2.2 インデント保持
- 元のコードのインデントスタイル（スペース/タブ）を保持
- 相対的なインデントレベルの計算と適用
- ネストされた構造の適切な処理

#### 2.3 エラー報告の強化
- コンテキスト行（前後のコード）を含むエラーメッセージ
- 類似度スコアの表示
- 最も近い一致の提示

### 3. 実装手順

1. `internal/proposal/diff.go`に新しい戦略を追加
   - 既存のDiffUtilityは維持しつつ、新しいSearchReplaceDiffStrategyを実装
   - インターフェースを定義して両方の戦略をサポート

2. 設定オプションの追加
   - fuzzyThreshold（類似度閾値）
   - bufferLines（コンテキスト行数）
   - 使用する戦略の選択

3. エラーハンドリングの強化
   - より詳細なエラー型の定義
   - コンテキスト情報の収集と表示

### 4. 使用例

```go
strategy := NewSearchReplaceDiffStrategy(0.9, 20) // 90%の類似度要求、20行のコンテキスト

result, err := strategy.ApplyDiff(
    originalContent,
    searchContent,
    replaceContent,
    startLine,
    endLine,
)
if err != nil {
    // 詳細なエラー情報を取得
    fmt.Printf("差分適用エラー: %v\n", err)
    return
}
```

## 期待される効果

1. より安全な差分適用
   - 誤った箇所への適用を防止
   - コードスタイルの保持

2. 開発者体験の向上
   - 詳細なエラー情報による問題の早期発見
   - より柔軟な差分適用オプション

3. メンテナンス性の向上
   - 明確な戦略パターンによる拡張性
   - テスト容易性の向上