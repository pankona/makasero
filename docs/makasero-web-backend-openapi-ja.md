# makasero Web Backend API 仕様

このドキュメントでは、makasero Web BackendのREST APIについて説明します。

## 概要

makasero Web Backend APIは、AIエージェントセッションを管理するためのインターフェースを提供します。
このAPIを使用して、新しいセッションの作成、既存セッションの状態確認、およびセッションへのコマンド送信が可能です。

### 基本URL

```
/api
```

## エンドポイント

### セッションの作成

新しいAIエージェントセッションを作成し、指定されたプロンプトの処理を開始します。

```
POST /api/sessions
```

#### リクエスト

```json
{
  "prompt": "AIエージェントへの指示"
}
```

| フィールド | 型 | 説明 |
|-----------|------|-------------|
| prompt | string | セッション開始時のユーザープロンプト |

#### レスポンス

```json
{
  "session_id": "12345678-1234-5678-1234-567812345678",
  "status": "accepted"
}
```

| フィールド | 型 | 説明 |
|-----------|------|-------------|
| session_id | string | 作成されたセッションの一意識別子 |
| status | string | セッションの状態（acceptedなど） |

#### ステータスコード

- `202 Accepted`: セッションが正常に作成され、処理が開始された
- `400 Bad Request`: 無効なリクエストパラメータ
- `500 Internal Server Error`: サーバー内部エラー

### セッション状態の取得

指定されたセッションIDのセッション状態を取得します。

```
GET /api/sessions/{sessionId}
```

#### パラメータ

| パラメータ | 型 | 説明 |
|-----------|------|-------------|
| sessionId | string | 取得するセッションのID |

#### レスポンス

```json
{
  "id": "12345678-1234-5678-1234-567812345678",
  "created_at": "2025-04-28T05:00:00Z",
  "updated_at": "2025-04-28T05:01:00Z",
  "serialized_history": [
    {
      "role": "user",
      "parts": [
        {
          "type": "text",
          "content": "最初のプロンプト"
        }
      ]
    },
    {
      "role": "model",
      "parts": [
        {
          "type": "text",
          "content": "AIの応答"
        }
      ]
    }
  ]
}
```

#### ステータスコード

- `200 OK`: セッション情報が正常に取得された
- `404 Not Found`: 指定されたセッションIDが見つからない
- `500 Internal Server Error`: サーバー内部エラー

### セッションへのコマンド送信

既存のセッションに新しいコマンドを送信します。

```
POST /api/sessions/{sessionId}/commands
```

#### パラメータ

| パラメータ | 型 | 説明 |
|-----------|------|-------------|
| sessionId | string | コマンドを送信するセッションのID |

#### リクエスト

```json
{
  "command": "AIエージェントへの追加コマンド"
}
```

| フィールド | 型 | 説明 |
|-----------|------|-------------|
| command | string | セッションに送信するコマンド |

#### レスポンス

```json
{
  "message": "Command accepted"
}
```

| フィールド | 型 | 説明 |
|-----------|------|-------------|
| message | string | コマンド受付状態のメッセージ |

#### ステータスコード

- `202 Accepted`: コマンドが正常に受け付けられた
- `400 Bad Request`: 無効なリクエストパラメータ
- `404 Not Found`: 指定されたセッションIDが見つからない
- `500 Internal Server Error`: サーバー内部エラー

## データモデル

### Session

| フィールド | 型 | 説明 |
|-----------|------|-------------|
| id | string | セッションの一意識別子 |
| created_at | string (date-time) | セッション作成日時 |
| updated_at | string (date-time) | セッション最終更新日時 |
| serialized_history | array | セッション履歴 |

### SerializableContent

| フィールド | 型 | 説明 |
|-----------|------|-------------|
| role | string | コンテンツの役割（user/model） |
| parts | array | コンテンツのパート配列 |

### SerializablePart

| フィールド | 型 | 説明 |
|-----------|------|-------------|
| type | string | パートのタイプ（例: text） |
| content | string | パートの内容 | 