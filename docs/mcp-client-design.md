# MCP Client Integration Design for makasero

## 概要

makasero に MCP (Model Control Protocol) クライアント機能を統合するための設計ドキュメントです。
シンプルさを重視し、main パッケージ内で完結する実装を目指します。
ローカル実行に適した `StdioMCPClient` を採用します。

## 構成

```
makasero/
└── main.go    # MCP クライアント機能を含むメインの実装
```

## 実装方針

### 1. MCP クライアントの直接利用

- `github.com/mark3labs/mcp-go/client` パッケージの `StdioMCPClient` を直接利用
- 不要なラッパーやインターフェースの抽象化を避ける
- 必要最小限の機能のみを実装

### 2. 主要な機能

```go
func main() {
    // MCP クライアントの初期化
    client, err := client.NewStdioMCPClient(
        "claude",      // サーバーコマンド
        []string{},    // 環境変数
        "mcp", "serve" // 引数
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 標準エラー出力のキャプチャ
    go io.Copy(os.Stderr, client.Stderr())

    // 初期化リクエストの送信
    initResult, err := client.Initialize(ctx, mcp.InitializeRequest{})
    if err != nil {
        log.Fatal(err)
    }

    // 利用可能なツールの取得と変換
    mcpTools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
    if err != nil {
        log.Fatal(err)
    }

    // ツール一覧を model.Tools に統合
    // MCP のツールは名前の prefix に "mcp_" をつけて区別する
    tools := convertMCPTools(mcpTools.Tools, "mcp_")
    model.Tools = append(model.Tools, tools...)

    // 通知ハンドラの設定
    client.OnNotification(handleNotification)

    // プロンプトの構築（ツール情報を含める）
    prompt := buildPromptRequest(model.Tools)

    // プロンプトの送信と結果の処理
    result, err := client.CallTool(ctx, prompt)
    if err != nil {
        log.Printf("error: %v", err)
        return
    }
    
    // 結果の処理
    handleResult(result)
}

// MCP のツール定義を model.Tool に変換
func convertMCPTools(mcpTools []mcp.Tool, prefix string) []model.Tool {
    tools := make([]model.Tool, len(mcpTools))
    for i, t := range mcpTools {
        tools[i] = model.Tool{
            Name:        prefix + t.Name,  // MCP のツールには prefix をつける
            Description: t.Description,
            Parameters:  convertParameters(t.Parameters),
        }
    }
    return tools
}

// function calling のルーティング
func handleFunctionCall(name string, params map[string]interface{}) (interface{}, error) {
    // MCP のツールかどうかを prefix で判断
    if strings.HasPrefix(name, "mcp_") {
        // MCP ツールの場合
        return callMCPTool(strings.TrimPrefix(name, "mcp_"), params)
    }
    
    // 自前の関数の場合
    return callLocalFunction(name, params)
}

// MCP ツールの呼び出し
func callMCPTool(name string, params map[string]interface{}) (interface{}, error) {
    req := mcp.CallToolRequest{
        Tool: name,
        Args: params,
    }
    return client.CallTool(ctx, req)
}

// 自前の関数の呼び出し
func callLocalFunction(name string, params map[string]interface{}) (interface{}, error) {
    switch name {
    case "list_dir":
        return handleListDir(params)
    case "read_file":
        return handleReadFile(params)
    // ... 他の自前の関数
    default:
        return nil, fmt.Errorf("unknown function: %s", name)
    }
}
```

### 3. Function Calling の振り分け方針

1. **ツール名による区別**
   - MCP のツールには prefix として "mcp_" をつける
   - prefix の有無で振り分け先を判断

2. **優先順位**
   - 自前の関数を優先
   - MCP のツールは補完的に利用

3. **実行の流れ**
   - Gemini からの function calling 要求を受け取る
   - ツール名を確認して振り分け
   - 適切な実行関数にルーティング

### 4. エラーハンドリング

- シンプルなエラーチェックとログ出力
- 標準エラー出力のキャプチャと処理
- クリーンアップの確実な実行

### 5. 設定

```go
const (
    mcpServerCmd  = "claude"
    mcpServerArg1 = "mcp"
    mcpServerArg2 = "serve"
    mcpToolPrefix = "mcp_"  // MCP ツールの prefix
)
```

## 実装ステップ

1. Phase 1: 基本実装 (1週間)
   - MCP クライアントの初期化と接続
   - 初期化処理とツール一覧の取得
   - ツール定義の変換と統合
   - function calling の振り分け実装
   - 標準エラー出力のハンドリング

2. Phase 2: 機能追加 (1週間)
   - ツール呼び出し機能の実装
   - 結果のハンドリング改善
   - テストの追加

## テスト方針

- main パッケージ内でのテスト
- 実際の MCP サーバープロセスとの結合テスト
- エラーケースの検証

## セキュリティ

- 環境変数での設定管理
- 基本的な入力バリデーション
- サブプロセスの適切な終了処理 