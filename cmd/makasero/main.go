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
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/option"
)

var (
	debug      = flag.Bool("debug", false, "デバッグモード")
	promptFile = flag.String("f", "", "プロンプトファイル")
)

const (
	mcpServerCmd  = "claude"
	mcpServerArg1 = "mcp"
	mcpServerArg2 = "serve"
	mcpToolPrefix = "mcp_"
)

var mcpClient *client.StdioMCPClient

func debugPrint(format string, args ...any) {
	if *debug {
		fmt.Printf("[DEBUG] "+format, args...)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func readPromptFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("プロンプトファイルの読み込みに失敗: %v", err)
	}
	return string(content), nil
}

func run() error {
	// コマンドライン引数の処理
	flag.Parse()

	debugPrint("MCP クライアントの初期化を開始します\n")
	// MCP クライアントの初期化
	var err error
	mcpClient, err = client.NewStdioMCPClient(
		mcpServerCmd,
		[]string{},
		mcpServerArg1,
		mcpServerArg2,
	)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %v", err)
	}
	//defer mcpClient.Close()

	debugPrint("標準エラー出力のキャプチャを開始します\n")
	// 標準エラー出力のキャプチャ
	go io.Copy(os.Stderr, mcpClient.Stderr())

	debugPrint("MCP クライアントの初期化リクエストを送信します\n")
	// 初期化リクエストの送信
	if _, err := mcpClient.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		return fmt.Errorf("failed to initialize MCP client: %v", err)
	}

	debugPrint("利用可能なツールの一覧を取得します\n")
	// 利用可能なツールの取得と変換
	mcpTools, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list MCP tools: %v", err)
	}
	debugPrint("取得したツール数: %d\n", len(mcpTools.Tools))

	debugPrint("MCP のツールを functions に追加します\n")
	debugPrint("functions マップの初期サイズ: %d\n", len(functions))

	// MCP のツールを functions に追加
	for _, t := range mcpTools.Tools {
		name := mcpToolPrefix + t.Name
		debugPrint("\n=== ツール %s の登録を開始 ===\n", name)
		debugPrint("元のツール情報:\n")
		debugPrint("  Name: %s\n", t.Name)
		debugPrint("  Description: %s\n", t.Description)
		debugPrint("  InputSchema: %+v\n", t.InputSchema)

		// 説明文を整形
		description := t.Description
		// 改行を空白に置換
		description = strings.ReplaceAll(description, "\n", " ")
		// 連続する空白を1つに
		description = strings.Join(strings.Fields(description), " ")
		debugPrint("説明文の処理:\n")
		debugPrint("  元の長さ: %d\n", len(t.Description))
		debugPrint("  処理後の長さ: %d\n", len(description))
		debugPrint("  処理後の内容: %s\n", description)

		// パラメータの変換
		convertedParams := convertMCPParameters(t.InputSchema)
		debugPrint("パラメータの変換結果:\n")
		for paramName, paramSchema := range convertedParams {
			debugPrint("  パラメータ [%s]:\n", paramName)
			debugPrint("    Type: %v\n", paramSchema.Type)
			debugPrint("    Description: %s\n", paramSchema.Description)
			if paramSchema.Items != nil {
				debugPrint("    Items: %+v\n", paramSchema.Items)
			}
		}

		// FunctionDeclaration の作成
		declaration := &genai.FunctionDeclaration{
			Name:        name,
			Description: description,
			Parameters: &genai.Schema{
				Type:       genai.TypeObject,
				Properties: convertedParams,
			},
		}
		if len(t.InputSchema.Required) > 0 {
			declaration.Parameters.Required = t.InputSchema.Required
			debugPrint("必須パラメータを設定: %v\n", t.InputSchema.Required)
		}

		// Handler の作成
		handler := func(ctx context.Context, args map[string]any) (map[string]any, error) {
			debugPrint("MCP ツール %s を呼び出します\n引数: %+v\n", t.Name, args)
			result, err := callMCPTool(t.Name, args)
			if err != nil {
				debugPrint("MCP ツール %s の呼び出しに失敗: %v\n", t.Name, err)
				return nil, err
			}
			debugPrint("MCP ツール %s の呼び出し結果: %+v\n", t.Name, result)

			// MCP の結果を map に変換
			mcpResult, ok := result.(*mcp.CallToolResult)
			if !ok {
				return nil, fmt.Errorf("unexpected result type: %T", result)
			}

			// 結果を文字列の配列に変換
			var contents []string
			for _, content := range mcpResult.Content {
				if textContent, ok := content.(mcp.TextContent); ok {
					contents = append(contents, textContent.Text)
				} else {
					contents = append(contents, fmt.Sprintf("%v", content))
				}
			}

			resultMap := map[string]any{
				"is_error": mcpResult.IsError,
				"content":  strings.Join(contents, "\n"),
			}
			if mcpResult.Result.Meta != nil {
				resultMap["meta"] = mcpResult.Result.Meta
			}

			return map[string]any{"result": resultMap}, nil
		}

		// functions マップに登録
		functions[name] = FunctionDefinition{
			Declaration: declaration,
			Handler:     handler,
		}
		debugPrint("ツール %s の登録が完了しました\n", name)
		debugPrint("functions マップの現在のサイズ: %d\n", len(functions))
	}

	debugPrint("\n=== 登録されたツール一覧 ===\n")
	for name := range functions {
		debugPrint("- %s\n", name)
	}

	debugPrint("通知ハンドラを設定します\n")
	// 通知ハンドラの設定
	mcpClient.OnNotification(handleNotification)
	debugPrint("通知ハンドラの設定が完了しました\n")

	// セッション一覧表示の処理
	if *listSessionsFlag {
		debugPrint("セッション一覧の表示を開始します\n")
		return listSessions()
	}

	// 会話履歴全文表示の処理
	if *showHistory != "" {
		debugPrint("セッション履歴の表示を開始します: %s\n", *showHistory)
		return showSessionHistory(*showHistory)
	}

	args := flag.Args()
	debugPrint("コマンドライン引数: %v\n", args)

	// プロンプトの取得
	var userInput string
	if *promptFile != "" {
		// ファイルからプロンプトを読み込む
		prompt, err := readPromptFromFile(*promptFile)
		if err != nil {
			return err
		}
		userInput = prompt
		fmt.Printf("プロンプトファイルから読み込んだ内容:\n%s\n", userInput)
	} else if len(args) > 0 {
		// コマンドライン引数からプロンプトを取得
		userInput = strings.Join(args, " ")
		fmt.Printf("コマンドライン引数から取得したプロンプト:\n%s\n", userInput)
	} else {
		return fmt.Errorf("プロンプトを指定してください（コマンドライン引数または -f オプション）")
	}

	// APIキーの取得
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY 環境変数が設定されていません")
	}

	// モデル名の取得（デフォルト: gemini-2.0-flash-lite）
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-2.0-flash-lite"
	}

	// コンテキストの作成
	ctx := context.Background()

	// クライアントの初期化
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("クライアントの初期化に失敗: %v", err)
	}
	defer client.Close()

	// モデルの初期化
	model := client.GenerativeModel(modelName)
	debugPrint("モデル %s を初期化しました\n", modelName)

	// 関数定義から FunctionDeclaration のスライスを作成
	var declarations []*genai.FunctionDeclaration
	debugPrint("登録されている関数の数: %d\n", len(functions))
	for name, fn := range functions {
		debugPrint("関数を登録: %s\n", name)
		debugPrint("関数の説明: %s\n", fn.Declaration.Description)
		if fn.Declaration.Parameters != nil {
			debugPrint("パラメータ: %+v\n", fn.Declaration.Parameters)
		}
		declarations = append(declarations, fn.Declaration)
	}

	// モデルに設定
	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: declarations,
		},
	}
	debugPrint("モデルにツールを設定しました\n")

	// セッションの読み込み
	var session *Session
	if *sessionID != "" {
		debugPrint("セッション %s を読み込みます\n", *sessionID)
		var err error
		session, err = loadSession(*sessionID)
		if err != nil {
			return err
		}
		debugPrint("セッションを読み込みました\n")
	} else {
		// 新規セッション
		session = &Session{
			ID:        generateSessionID(),
			CreatedAt: time.Now(),
		}
		debugPrint("新規セッション %s を作成しました\n", session.ID)
	}

	// チャットセッションを開始
	chat := model.StartChat()
	debugPrint("チャットセッションを開始しました\n")
	if len(session.History) > 0 {
		chat.History = session.History
		debugPrint("履歴を読み込みました（%d 件）\n", len(session.History))
	}

	// メッセージの送信と応答の取得
	fmt.Printf("\nAIに送信するメッセージ:\n%s\n\n", userInput)
	debugPrint("メッセージを送信します\n")
	resp, err := chat.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		debugPrint("メッセージの送信に失敗: %v\n", err)
		debugPrint("レスポンスの詳細:\n")
		if resp != nil {
			debugPrint("  Candidates: %+v\n", resp.Candidates)
			debugPrint("  PromptFeedback: %+v\n", resp.PromptFeedback)
			if len(resp.Candidates) > 0 {
				debugPrint("  最初の候補の詳細:\n")
				debugPrint("    FinishReason: %v\n", resp.Candidates[0].FinishReason)
				debugPrint("    SafetyRatings: %+v\n", resp.Candidates[0].SafetyRatings)
				debugPrint("    CitationMetadata: %+v\n", resp.Candidates[0].CitationMetadata)
				if resp.Candidates[0].Content != nil {
					debugPrint("    Content.Parts: %+v\n", resp.Candidates[0].Content.Parts)
					debugPrint("    Content.Role: %v\n", resp.Candidates[0].Content.Role)
				}
			}
		} else {
			debugPrint("  レスポンスが nil です\n")
		}
		// エラーが発生しても、それまでの履歴は保存
		session.History = chat.History
		saveSession(session)
		return fmt.Errorf("メッセージの送信に失敗: %v", err)
	}
	debugPrint("メッセージを送信しました\n")

	var shouldBreak bool
	for !shouldBreak {
		shouldBreak = true

		// レスポンスの処理
		if len(resp.Candidates) > 0 {
			debugPrint("応答候補数: %d\n", len(resp.Candidates))
			cand := resp.Candidates[0]
			if cand.Content != nil {
				debugPrint("応答パーツ数: %d\n", len(cand.Content.Parts))
				for i, part := range cand.Content.Parts {
					debugPrint("パーツ %d の型: %T\n", i, part)
					switch p := part.(type) {
					case genai.FunctionCall:
						// 関数呼び出しの場合
						debugPrint("関数呼び出し: %s\n", p.Name)
						debugPrint("引数: %+v\n", p.Args)

						if p.Name == "complete" || p.Name == "askQuestion" {
							debugPrint("完了関数が呼び出されました\n")
							session.History = chat.History
							session.UpdatedAt = time.Now()
							if err := saveSession(session); err != nil {
								return err
							}
							fmt.Printf("\nセッションID: %s\n", session.ID)
							return nil
						}

						fn, exists := functions[p.Name]
						if !exists {
							debugPrint("未知の関数が呼び出されました: %s\n", p.Name)
							return fmt.Errorf("unknown function: %s", p.Name)
						}

						debugPrint("関数 %s を実行します\n", p.Name)
						result, err := fn.Handler(ctx, p.Args)
						if err != nil {
							debugPrint("関数の実行に失敗: %v\n", err)
							return fmt.Errorf("function %s failed: %v", p.Name, err)
						}
						debugPrint("関数の実行結果: %+v\n", result)

						// 実行結果を FunctionResponse として送信
						debugPrint("実行結果を送信します\n")
						resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
							Name:     p.Name,
							Response: result,
						})
						if err != nil {
							debugPrint("実行結果の送信に失敗: %v\n", err)
							return fmt.Errorf("実行結果の送信に失敗: %v", err)
						}
						debugPrint("実行結果を送信しました\n")

						// complete 関数以外の場合は続きのタスクを実行するために、ループを継続
						shouldBreak = false
					case genai.Text:
						// テキスト応答の場合
						debugPrint("テキスト応答:\n%s\n", p)
						fmt.Printf("\nAIからの応答:\n%s\n", p)
					default:
						debugPrint("未知の応答タイプ: %T\n", part)
					}
				}
			} else {
				debugPrint("応答の Content が nil です\n")
			}
		} else {
			debugPrint("応答候補がありません\n")
		}
	}

	debugPrint("セッションを保存します\n")
	session.History = chat.History
	session.UpdatedAt = time.Now()
	if err := saveSession(session); err != nil {
		return err
	}
	fmt.Printf("\nセッションID: %s\n", session.ID)

	return nil
}

// MCP ツールの呼び出し
func callMCPTool(name string, args map[string]any) (interface{}, error) {
	debugPrint("callMCPTool: name=%s, args=%+v\n", name, args)
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	debugPrint("リクエスト: %+v\n", req)
	result, err := mcpClient.CallTool(context.Background(), req)
	if err != nil {
		debugPrint("ツール呼び出しエラー: %v\n", err)
		return nil, err
	}
	debugPrint("ツール呼び出し結果: %+v\n", result)
	return result, nil
}

// マップのキーを取得するヘルパー関数
func getMapKeys(m interface{}) []string {
	var keys []string
	switch v := m.(type) {
	case map[string]interface{}:
		for k := range v {
			keys = append(keys, k)
		}
	case map[interface{}]interface{}:
		for k := range v {
			if str, ok := k.(string); ok {
				keys = append(keys, str)
			}
		}
	}
	return keys
}

// MCP のパラメータを Gemini の FunctionDeclaration のパラメータに変換
func convertMCPParameters(schema mcp.ToolInputSchema) map[string]*genai.Schema {
	debugPrint("パラメータ変換開始: %+v\n", schema)

	converted := make(map[string]*genai.Schema)
	if schema.Properties == nil {
		debugPrint("Properties が nil です\n")
		return converted
	}

	for name, p := range schema.Properties {
		debugPrint("\n--- プロパティ %s の処理開始 ---\n", name)
		prop, ok := p.(map[string]interface{})
		if !ok {
			debugPrint("プロパティの型変換に失敗: %s (実際の型: %T)\n", name, p)
			continue
		}

		// type フィールドの確認
		typeVal, ok := prop["type"].(string)
		if !ok {
			debugPrint("type フィールドの取得に失敗: %v\n", prop["type"])
			continue
		}

		// description フィールドの確認（オプション）
		description := ""
		if desc, ok := prop["description"].(string); ok {
			description = desc
		}

		// スキーマの作成
		schema := &genai.Schema{
			Type:        convertSchemaType(typeVal),
			Description: description,
		}

		// 配列型の場合は Items を設定
		if typeVal == "array" {
			if items, ok := prop["items"].(map[string]interface{}); ok {
				itemType, hasType := items["type"].(string)
				itemDesc, hasDesc := items["description"].(string)
				if hasType {
					schema.Items = &genai.Schema{
						Type: convertSchemaType(itemType),
					}
					if hasDesc {
						schema.Items.Description = itemDesc
					}
				}
			}
		}

		// オブジェクト型の場合は Properties を設定
		if typeVal == "object" {
			if properties, ok := prop["properties"].(map[string]interface{}); ok {
				schema.Properties = make(map[string]*genai.Schema)
				for subName, subProp := range properties {
					if subPropMap, ok := subProp.(map[string]interface{}); ok {
						subType, hasType := subPropMap["type"].(string)
						subDesc, hasDesc := subPropMap["description"].(string)
						if hasType {
							schema.Properties[subName] = &genai.Schema{
								Type: convertSchemaType(subType),
							}
							if hasDesc {
								schema.Properties[subName].Description = subDesc
							}
						}
					}
				}
			}
			// Required フィールドの設定
			if required, ok := prop["required"].([]interface{}); ok {
				schema.Required = make([]string, 0, len(required))
				for _, r := range required {
					if str, ok := r.(string); ok {
						schema.Required = append(schema.Required, str)
					}
				}
			}
		}

		converted[name] = schema
		debugPrint("プロパティ %s の変換が完了: %+v\n", name, schema)
	}

	return converted
}

// JSON Schema の型を Gemini の型に変換
func convertSchemaType(schemaType string) genai.Type {
	switch schemaType {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeString // デフォルトは string
	}
}

// 通知ハンドラ
func handleNotification(notification mcp.JSONRPCNotification) {
	debugPrint("Received notification: %v\n", notification)
}
