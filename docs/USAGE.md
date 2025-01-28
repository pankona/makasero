# Makasero CLI使用例

## 基本的な使い方

### コードの説明（explain）

1. 直接コードを指定する場合：
```bash
./bin/makasero explain "fmt.Println('Hello, World!')"
```

2. ファイルからコードを抽出して説明する場合：
```bash
# 特定の関数を抽出して説明
sed -n '/^func executeChat/,/^}/p' cmd/makasero/main.go | ./bin/makasero explain "$(cat)"

# 複数の関数を一度に抽出して説明
sed -n '/^func Execute/,/^}/p' main.go | ./bin/makasero explain "$(cat)"
```

### チャット（chat）

1. 単純な質問：
```bash
./bin/makasero chat "Goの並行処理について教えてください"
```

2. コードの改善提案：
```bash
# ファイルの内容を渡して改善提案を要求
cat path/to/file.go | ./bin/makasero chat "このコードを改善してください"

# 特定の要件での改善を要求
cat main.go | ./bin/makasero chat "このコードを以下の要件で改善してください：
1. エラーハンドリングの追加
2. ログ出力の実装
3. テストのしやすさの向上"
```

3. パイプラインでの使用：
```bash
# grepで特定のパターンを抽出して改善
grep -r "TODO" . | ./bin/makasero chat "これらのTODOに対する実装案を提案してください"

# gitの差分を説明
git diff | ./bin/makasero chat "この変更内容をレビューしてください"
```

## 応用例

### コードレビュー支援

1. ファイルの改善提案：
```bash
# 単一ファイルの改善
cat main.go | ./bin/makasero chat "このコードをより良いGoの慣習に従うように改善してください"

# 複数ファイルの改善
find . -name "*.go" -exec sh -c 'echo "=== {} ==="; cat {}' \; | \
./bin/makasero chat "これらのファイルのコーディング規約違反を指摘し、修正案を提示してください"
```

2. PRのレビュー支援：
```bash
# PRの差分をレビュー
git diff origin/main...HEAD | ./bin/makasero chat "この変更に対するレビューコメントを生成してください。
以下の点に注目してください：
1. コーディング規約への準拠
2. パフォーマンスへの影響
3. セキュリティの考慮"
```

### バッチ処理での使用

1. 大規模なリファクタリング：
```bash
# 特定のパターンを含むファイルを検索して改善提案を生成
find . -type f -name "*.go" -exec grep -l "deprecated" {} \; | \
while read file; do
  echo "=== $file ==="
  cat "$file" | ./bin/makasero chat "このファイルの非推奨APIの使用を最新のAPIに更新してください"
done
```

2. ドキュメント生成：
```bash
# 関数のドキュメントを生成
find . -name "*.go" -exec sh -c 'echo "=== {} ==="; cat {}' \; | \
./bin/makasero chat "これらの関数に対するGoDocコメントを生成してください"
```

## Tips

### バックアップディレクトリの指定

コードの変更を適用する際のバックアップ先を指定できます：
```bash
./bin/makasero chat --backup-dir=/path/to/backup "コードを改善してください"
```

### エイリアスの設定

~/.bashrcや~/.zshrcに以下のようなエイリアスを設定すると便利です：

```bash
# コードの説明
alias explain='./bin/makasero explain'
# チャットと改善提案
alias chat='./bin/makasero chat'
```

使用例：
```bash
explain "$(cat main.go)"
cat main.go | chat "このコードを改善してください"
```

### 実行例

1. 基本的な改善提案：
```bash
$ cat main.go | ./bin/makasero chat "このコードを改善してください"
提案内容：
エラーハンドリングを改善し、ログ出力を追加します。

対象ファイル：main.go
変更内容：
@@ -10,6 +10,7 @@
 func main() {
-    result := process()
+    result, err := process()
+    if err != nil {
+        log.Printf("処理エラー: %v", err)
+        return fmt.Errorf("実行エラー: %w", err)
+    }
     return nil
 }

この提案を適用しますか？ [y/N]:
```

2. 特定の要件での改善：
```bash
$ cat sample.go | ./bin/makasero chat "ユーザー入力の処理を改善してください"
提案内容：
入力のバリデーションとエラーハンドリングを追加します。

対象ファイル：sample.go
変更内容：
@@ -5,7 +5,12 @@
 func main() {
     var input string
-    fmt.Scanln(&input)
+    fmt.Print("Enter value: ")
+    if _, err := fmt.Scanln(&input); err != nil {
+        fmt.Fprintf(os.Stderr, "入力エラー: %v\n", err)
+        os.Exit(1)
+    }
+    if input == "" {
+        fmt.Fprintln(os.Stderr, "入力が必要です")
+        os.Exit(1)
+    }
     fmt.Printf("入力値: %s\n", input)
 }

この提案を適用しますか？ [y/N]: