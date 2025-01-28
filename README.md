# Makasero CLI

Makasero CLIは、OpenAI GPT-4を活用したコードの説明や対話を行うためのコマンドラインツールです。

## 機能

- コードの説明（explain）: 指定したコードの説明を生成
- チャット（chat）: GPT-4との対話機能

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

```bash
makasero -command explain -input "func hello() { fmt.Println('Hello, World!') }"
```

### チャット

```bash
makasero -command chat -input '[{"role":"user","content":"Goでの並行処理について説明してください"}]'
```

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