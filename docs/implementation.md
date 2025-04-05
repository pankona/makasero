# makasero 実装詳細

## 依存関係

```go
require (
    github.com/google/generative-ai-go v0.5.0
    github.com/spf13/cobra v1.8.0
)
```

## 実装詳細

### Tool インターフェース

`tools/tool.go` に以下のインターフェースを定義します：

```go
package tools

type Tool interface {
    // ツールの名前を返す
    Name() string
    
    // ツールの説明を返す
    Description() string
    
    // ツールの実行
    Execute(args map[string]interface{}) (string, error)
}
```

### execCommand ツール

`tools/exec.go` に以下の実装を行います：

```go
package tools

type ExecCommand struct{}

func (e *ExecCommand) Name() string {
    return "execCommand"
}

func (e *ExecCommand) Description() string {
    return "Execute a shell command and return its output"
}

func (e *ExecCommand) Execute(args map[string]interface{}) (string, error) {
    cmd, ok := args["command"].(string)
    if !ok {
        return "", fmt.Errorf("command argument is required")
    }
    
    output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("command execution failed: %v", err)
    }
    
    return string(output), nil
}
```

### Agent 実装

`agent/agent.go` に以下の実装を行います：

```go
package agent

type Agent struct {
    client    *gemini.Client
    tools     map[string]tools.Tool
}

func New(apiKey string) (*Agent, error) {
    client, err := gemini.NewClient(apiKey)
    if err != nil {
        return nil, err
    }
    
    return &Agent{
        client: client,
        tools:  make(map[string]tools.Tool),
    }, nil
}

func (a *Agent) RegisterTool(tool tools.Tool) {
    a.tools[tool.Name()] = tool
}

func (a *Agent) Process(input string) (string, error) {
    // Geminiにプロンプトを送信し、ツールの使用を判断
    // ツールの実行が必要な場合は、適切なツールを呼び出し
    // 結果を返す
}
```

### CLI 実装

`cmd/makasero/main.go` に以下の実装を行います：

```go
package main

func main() {
    var apiKey string
    
    rootCmd := &cobra.Command{
        Use:   "makasero",
        Short: "AI agent CLI tool",
        RunE: func(cmd *cobra.Command, args []string) error {
            agent, err := agent.New(apiKey)
            if err != nil {
                return err
            }
            
            // execCommand ツールを登録
            agent.RegisterTool(&tools.ExecCommand{})
            
            // 対話モードの実装
            return nil
        },
    }
    
    rootCmd.Flags().StringVar(&apiKey, "api-key", "", "Gemini API key")
    rootCmd.MarkFlagRequired("api-key")
    
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## 使用方法

```bash
# ビルド
go build -o makasero cmd/makasero/main.go

# 実行
./makasero --api-key YOUR_GEMINI_API_KEY
```

## 今後の拡張ポイント

1. ツールの追加
   - 新しいツールを実装し、Agentに登録するだけで追加可能

2. LLMプロバイダーの追加
   - 新しいLLMプロバイダーをサポートするには、新しいクライアントを実装

3. 設定の外部化
   - 設定ファイルのサポート
   - 環境変数による設定

4. プラグインシステム
   - 動的ロード可能なツール
   - カスタムプロバイダーのサポート 