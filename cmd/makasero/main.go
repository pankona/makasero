package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/samber/lo"
	"google.golang.org/api/option"
)

var (
	debug          = flag.Bool("debug", false, "debug mode")
	promptFile     = flag.String("f", "", "prompt file")
	configFilePath = flag.String("config", "", "path to config file")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func readPromptFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %v", err)
	}
	return string(content), nil
}

func run() error {
	// コマンドライン引数の処理
	flag.Parse()

	config, err := LoadConfig(*configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load config: %v\nPlease create a config file at ~/.makasero/config.json with your MCP server settings", err)
	}

	mcpManager := NewMCPClientManager()
	ctx := context.Background()

	if err := mcpManager.InitializeFromConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize MCP clients: %v", err)
	}
	// いったん無効化する。MCP Server プロセスをキルする必要があるが今はそういう動きをしてくれないっぽい

	// 標準エラー出力のキャプチャ
	stderrReaders := mcpManager.GetStderrReaders()
	for serverName, reader := range stderrReaders {
		serverNameCopy := serverName
		go func(r io.Reader) {
			if *debug {
				buf := make([]byte, 1024)
				for {
					n, err := r.Read(buf)
					if err != nil {
						if err != io.EOF {
							fmt.Fprintf(os.Stderr, "[%s] stderr read error: %v\n", serverNameCopy, err)
						}
						return
					}
					fmt.Fprintf(os.Stderr, "[%s] %s", serverNameCopy, buf[:n])
				}
			} else {
				io.Copy(os.Stderr, r)
			}
		}(reader)
	}

	// 通知ハンドラの設定
	mcpManager.SetupNotificationHandlers(func(serverName string, notification mcp.JSONRPCNotification) {
		if *debug {
			fmt.Printf("[%s] Notification: %v\n", serverName, notification)
		} else {
			handleNotification(notification)
		}
	})

	// 利用可能なツールの取得と変換
	mcpFuncDecls, err := mcpManager.GenerateAllFunctionDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate MCP tools: %v", err)
	}

	// セッション一覧表示の処理
	if *listSessionsFlag {
		return listSessions()
	}

	// 会話履歴全文表示の処理
	if *showHistory != "" {
		return showSessionHistory(*showHistory)
	}

	args := flag.Args()

	// プロンプトの取得
	var userInput string
	if *promptFile != "" {
		// ファイルからプロンプトを読み込む
		prompt, err := readPromptFromFile(*promptFile)
		if err != nil {
			return err
		}
		userInput = prompt
	} else if len(args) > 0 {
		// コマンドライン引数からプロンプトを取得
		userInput = strings.Join(args, " ")
	} else {
		return fmt.Errorf("Please specify a prompt (command line arguments or -f option)")
	}

	// APIキーの取得
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	// モデル名の取得（デフォルト: gemini-2.0-flash-lite）
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-2.0-flash-lite"
	}

	// コンテキストの作成

	// クライアントの初期化
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("failed to initialize client: %v", err)
	}
	defer client.Close()

	// モデルの初期化
	model := client.GenerativeModel(modelName)

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("あなたはAIアシスタントです。ユーザーからのタスクを実行し、タスクが完了したら必ず「complete」関数を呼び出してください。関数を呼び出す際は、テキストで関数名を書くのではなく、実際に関数を呼び出してください。"),
		},
	}

	functions := make(map[string]FunctionDefinition)

	// 関数定義から FunctionDeclaration のスライスを作成
	for _, fn := range mcpFuncDecls {
		functions[fn.Declaration.Name] = fn
	}
	for _, fn := range myFunctions {
		functions[fn.Declaration.Name] = fn
	}

	// モデルに function calling 設定
	mcpFuncDeclarations := lo.Map(mcpFuncDecls, func(fn FunctionDefinition, _ int) *genai.FunctionDeclaration {
		return fn.Declaration
	})

	var allFuncDeclarations []*genai.FunctionDeclaration
	allFuncDeclarations = append(allFuncDeclarations, mcpFuncDeclarations...)

	for _, fn := range functions {
		allFuncDeclarations = append(allFuncDeclarations, fn.Declaration)
	}

	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: allFuncDeclarations,
		},
	}

	model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingAuto,
		},
	}

	// list tools
	fmt.Printf("declared tools: %d\n", len(functions))
	for _, tool := range functions {
		fmt.Printf("%s\n", tool.Declaration.Name)
	}

	// セッションの読み込み
	var session *Session
	if *sessionID != "" {
		var err error
		session, err = loadSession(*sessionID)
		if err != nil {
			return err
		}
	} else {
		// 新規セッション
		session = &Session{
			ID:        generateSessionID(),
			CreatedAt: time.Now(),
		}
	}

	// チャットセッションを開始
	chat := model.StartChat()
	if len(session.History) > 0 {
		chat.History = session.History
	}

	fmt.Println("\n--- Start session ---")

	// メッセージの送信と応答の取得
	fmt.Printf("\n🗣️ Sending message to AI:\n%s\n", strings.TrimSpace(userInput))

	resp, err := chat.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		// エラーが発生しても、それまでの履歴は保存
		session.History = chat.History
		saveSession(session)
		return fmt.Errorf("failed to send message to AI: %v", err)
	}

	var shouldBreak bool
	for !shouldBreak {
		shouldBreak = true

		// レスポンスの処理
		if len(resp.Candidates) > 0 {
			cand := resp.Candidates[0]
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					switch p := part.(type) {
					case genai.FunctionCall:
						fmt.Printf("\n🔧 AI uses function calling: %s\n", p.Name)

						// 関数呼び出しの場合
						if p.Name == "complete" || p.Name == "askQuestion" {
							session.History = chat.History
							session.UpdatedAt = time.Now()
							if err := saveSession(session); err != nil {
								return err
							}
							fmt.Printf("Session ID: %s\n", session.ID)
							return nil
						}

						if strings.HasPrefix(p.Name, "mcp_") {
							parts := strings.SplitN(p.Name, "_", 2)
							if len(parts) != 2 {
								return fmt.Errorf("invalid MCP tool name format: %s", p.Name)
							}

							result, err := mcpManager.CallMCPTool(ctx, p.Name, p.Args)
							if err != nil {
								return fmt.Errorf("MCP function %s failed: %v", p.Name, err)
							}

							// 実行結果を FunctionResponse として送信
							resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
								Name:     p.Name,
								Response: result,
							})
							if err != nil {
								return fmt.Errorf("failed to send function response: %v", err)
							}

							shouldBreak = false
							continue
						}

						fn, exists := functions[p.Name]
						if !exists {
							return fmt.Errorf("unknown function: %s", p.Name)
						}

						result, err := fn.Handler(ctx, p.Args)
						if err != nil {
							return fmt.Errorf("function %s failed: %v", p.Name, err)
						}

						// 実行結果を FunctionResponse として送信
						resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
							Name:     p.Name,
							Response: result,
						})
						if err != nil {
							return fmt.Errorf("failed to send function response: %v", err)
						}

						// complete 関数以外の場合は続きのタスクを実行するために、ループを継続
						shouldBreak = false
					case genai.Text:
						// テキスト応答の場合
						fmt.Printf("\n🤖 Response from AI:\n%s\n", strings.TrimSpace(string(p)))
					default:
						fmt.Printf("unknown response type: %T\n", part)
					}
				}
			} else {
				fmt.Printf("response content is nil\n")
			}
		} else {
			fmt.Printf("no response candidates\n")
		}
	}

	fmt.Println("\n--- Finish session ---")

	fmt.Printf("Saving session\n")
	session.History = chat.History
	session.UpdatedAt = time.Now()
	if err := saveSession(session); err != nil {
		return err
	}
	fmt.Printf("Session ID: %s\n", session.ID)

	return nil
}

// 通知ハンドラ
// TODO: まともに実装する
func handleNotification(notification mcp.JSONRPCNotification) {
	fmt.Printf("Received notification: %v\n", notification)
}
