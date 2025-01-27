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
```

使用例：
```bash
explain "$(cat main.go)"
chat '[{"role":"user","content":"Goのインターフェースについて説明してください"}]'