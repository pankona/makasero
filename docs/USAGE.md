# Roo CLI使用例

## 基本的な使い方

### コードの説明（explain）

1. 直接コードを指定する場合：
```bash
./bin/roo -command explain -input "fmt.Println('Hello, World!')"
```

2. ファイルからコードを抽出して説明する場合：
```bash
# 特定の関数を抽出して説明
sed -n '/^func executeChat/,/^}/p' cmd/roo/main.go | ./bin/roo -command explain -input "$(cat)"

# 複数の関数を一度に抽出して説明
sed -n '/^func Execute/,/^}/p' main.go | ./bin/roo -command explain -input "$(cat)"
```

### チャット（chat）

1. 単一のメッセージ：
```bash
./bin/roo -command chat -input '[{"role":"user","content":"こんにちは"}]'
```

2. 複数のメッセージを含む会話：
```bash
./bin/roo -command chat -input '[
  {"role":"system","content":"あなたはプログラミングの先生です"},
  {"role":"user","content":"Goの並行処理について教えてください"}
]'
```

### コード提案（propose）

1. パッチモードでの提案：
```bash
# 既存のコードに対して部分的な変更を提案
./bin/roo -command propose -input "path/to/file.go" -mode patch
```

2. 全体書き換えモードでの提案：
```bash
# コード全体の書き換えを提案
./bin/roo -command propose -input "path/to/file.go" -mode full
```

提案された変更は、差分形式で表示され、ユーザーの承認を得てから適用されます。

## 応用例

### ソースコードの解析

1. 特定のパターンにマッチする関数を検索して説明：
```bash
# mainで始まる関数を検索して説明
grep -A 10 "^func main" main.go | ./bin/roo -command explain -input "$(cat)"
```

2. 複数ファイルから関数を検索して説明：
```bash
# 全てのGoファイルからExecuteで始まる関数を検索
find . -name "*.go" -exec sh -c 'echo "=== {} ==="; sed -n "/^func Execute/,/^}/p" {}' \; | \
./bin/roo -command explain -input "$(cat)"
```

### コードレビュー支援

1. gitの差分を説明：
```bash
# 特定のコミットの変更内容を説明
git show <commit-hash> | ./bin/roo -command explain -input "$(cat)"

# 作業中の変更内容を説明
git diff | ./bin/roo -command explain -input "$(cat)"
```

2. PRのレビューコメント生成：
```bash
# PRの差分を説明してレビューコメントを生成
git diff origin/main...HEAD | ./bin/roo -command chat -input '[
  {"role":"system","content":"あなたはコードレビュアーです。以下の差分に対するレビューコメントを生成してください。"},
  {"role":"user","content":"'"$(cat)"'"}
]'
```

3. コードの改善提案：
```bash
# 特定のファイルに対して改善提案を生成
./bin/roo -command propose -input "main.go" -mode patch

# ディレクトリ内の全てのGoファイルに対して改善提案を生成
find . -name "*.go" -exec ./bin/roo -command propose -input {} -mode patch \;
```

## Tips

### 出力のフォーマット

- jqを使用してJSON出力を整形：
```bash
./bin/roo -command explain -input "fmt.Println('Hello')" | jq .data
```

### エイリアスの設定

~/.bashrcや~/.zshrcに以下のようなエイリアスを設定すると便利です：

```bash
# コードの説明
alias explain='./bin/roo -command explain -input'
# チャット
alias chat='./bin/roo -command chat -input'
# コード提案
alias propose='./bin/roo -command propose -input'
```

使用例：
```bash
explain "$(cat main.go)"
chat '[{"role":"user","content":"Goのインターフェースについて説明してください"}]'
propose "main.go" -mode patch