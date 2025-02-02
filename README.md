# Makasero CLI

Makasero CLIは、OpenAI GPT-4を活用したコードの説明や対話を行うためのコマンドラインツールです。

## 機能

- コードの説明（explain）: 指定したコードやファイルの説明を生成
- インタラクティブチャット（chat）: GPT-4との対話機能
  - ファイル編集モード: 指定したファイルの改善提案と自動適用
  - セッション管理: 会話履歴の保存と再開
  - バックアップ機能: 変更前のファイルを自動保存

## インストール

```bash
go install github.com/pankona/makasero/cmd/makasero@latest
```

## 使用方法

### 環境変数の設定

```bash
export OPENAI_API_KEY="your-api-key"
```

### コードの説明

ファイルの説明:
```bash
makasero explain path/to/file.go
```

コードスニペットの説明:
```bash
makasero explain "func hello() { fmt.Println('Hello, World!') }"
```

### チャット

通常のチャット:
```bash
makasero chat
```

ファイル編集モード:
```bash
makasero chat -f path/to/file.go
```

セッション管理:
```bash
# セッション一覧の表示
makasero chat -l

# セッションの再開
makasero chat -r <session-id>
```

## 対話コマンド

チャット中で使用できるコマンド:
- `exit` または `quit`: チャットを終了
- `/test`: テストの実行を提案（自動的にコマンドを生成）
- 通常の入力: AIとの対話
- コード改善の提案時: 変更の確認と適用（y/N）

## 特殊コマンド

チャット中で使用できる特殊コマンド:
- `/test`: テストの実行を提案
  - 現在のコンテキストに基づいて適切なテストコマンドを生成
  - 生成されたコマンドの確認と実行

## ディレクトリ構造

- `~/.makasero/sessions/`: セッション履歴の保存先
- `backups/`: ファイル変更時のバックアップ保存先（デフォルト）

## レスポンス形式

すべてのコマンドは以下の形式でJSONレスポンスを返します：

```json
{
  "success": true,
  "data": "レスポンスの内容",
  "error": null
}
```

エラーの場合：

```json
{
  "success": false,
  "data": null,
  "error": "エラーメッセージ"
}
```

## 詳細な使用例

より詳細な使用例については、[USAGE.md](docs/USAGE.md)を参照してください。sedやgrepを使用したコードの抽出方法や、gitとの連携など、実践的な使用例を紹介しています。

## ライセンス

MIT License