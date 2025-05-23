# makasero

AIエージェントを実装したコマンドラインツールです。

## 環境変数

- `GEMINI_API_KEY`: Gemini APIのキーを設定してください

## コマンドラインオプション

- `-debug`: デバッグモードを有効にする
- `-f`: プロンプトファイルのパスを指定
- `-ls`: 利用可能なセッション一覧を表示
- `-s`: 継続するセッションIDを指定（存在しないIDを指定すると新規セッションを開始）
- `-sh`: 指定したセッションIDの会話履歴全文を表示

## 実行例

プロンプトファイルから実行：
```bash
$ go run ./cmd/makasero -f prompt.txt
プロンプトファイルから読み込んだ内容:
カレントディレクトリの変更をコミットして。まだ add されていないものもあるから留意してね。
```

セッション一覧の表示：
```bash
$ go run ./cmd/makasero -ls
Session ID: 20250407222639_5f459f5e
Created: 2025-04-07T22:26:39Z
Messages: 8
初期プロンプト: カレントディレクトリの変更をコミットして...
```

新規セッションの開始：
```bash
$ go run ./cmd/makasero -s my_custom_session_id "Hello, AI"
新しいセッションを開始します。セッションID: my_custom_session_id
...
```

セッション履歴の表示：
```bash
$ go run ./cmd/makasero -sh 20250407222639_5f459f5e
セッションID: 20250407222639_5f459f5e
作成日時: 2025-04-07T22:26:39Z
最終更新: 2025-04-07T22:26:39Z
メッセージ数: 8

--- メッセージ 1 ---
役割: user
カレントディレクトリの変更をコミットして。まだ add されていないものもあるから留意してね。

--- メッセージ 2 ---
役割: model
はい、承知いたしました。
まず、カレントディレクトリの全ての変更をステージングエリアに追加します。次に、ステージングされた変更の差分を確認して、適切なコミットメッセージを作成し、コミットを実行します。
...
```

## GitHub Action: Interact with Makasero via Issue Comments

This repository includes a GitHub Action that allows you to interact with the `makasero` CLI by commenting on issues.

### How to Trigger

To trigger the action, simply add a comment to any issue in the repository with the following format:

```
/makasero <arguments>
```

Replace `<arguments>` with any valid command-line arguments you would normally pass to the `makasero` CLI.

### What it Does

When triggered, the GitHub Action will:
1. Build the `makasero` CLI.
2. Extract the issue description (the main body of the issue).
3. Execute `makasero` using the issue description as the primary input (similar to using the `-f` flag with a file containing the issue body) and the `<arguments>` you provided in the comment.

### How Makasero Responds

After `makasero` finishes processing, the action will post a new comment on the same issue. This comment will contain the output generated by `makasero`.

### Prerequisite: `GEMINI_API_KEY`

For the action to function correctly, you **must** configure the `GEMINI_API_KEY` as a secret in your repository's settings.
1. Go to your repository on GitHub.
2. Click on "Settings".
3. In the left sidebar, navigate to "Secrets and variables" > "Actions".
4. Click "New repository secret".
5. Name the secret `GEMINI_API_KEY`.
6. Paste your Gemini API key into the "Value" field.
7. Click "Add secret".

### Example

**Issue Title:** Refactor the authentication module

**Issue Body:**
The current authentication module is hard to maintain and lacks proper test coverage. We need to refactor it to improve modularity and add comprehensive unit tests. The focus should be on simplifying the login flow and ensuring error handling is robust.

**Your Comment:**
```
/makasero -debug Please suggest a plan to refactor the module described in the issue.
```

The action will then run `makasero` with the issue body and the argument `-debug "Please suggest a plan to refactor the module described in the issue."`. `makasero`'s output, including any generated plan or suggestions, will be posted as a new comment.

## Author

[@pankona](https://github.com/pankona)

## License

MIT License