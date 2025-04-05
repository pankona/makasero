package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"github.com/pankona/makasero/tools"
	"google.golang.org/api/option"
)

// GeminiClient はGeminiとの通信を管理するクライアントです
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
	chat   *genai.ChatSession
	tools  map[string]tools.Tool
}

// logResponse はレスポンスの内容をログ出力します
func logResponse(resp *genai.GenerateContentResponse) {
	log.Printf("Response candidates: %d\n", len(resp.Candidates))
	for i, candidate := range resp.Candidates {
		log.Printf("Candidate %d:\n", i)
		log.Printf("  FinishReason: %s\n", candidate.FinishReason)
		for j, part := range candidate.Content.Parts {
			log.Printf("  Part %d type: %T\n", j, part)
			switch v := part.(type) {
			case genai.Text:
				log.Printf("  Text: %s\n", string(v))
			case genai.FunctionCall:
				log.Printf("  FunctionCall: %s\n", v.Name)
				log.Printf("  Arguments: %v\n", v.Args)
			default:
				log.Printf("  Unknown part type: %T with value: %v\n", v, v)
			}
		}
	}
}

// NewGeminiClient は新しいGeminiClientを作成します
func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	// gemini-2.5-pro-exp-03-25モデルを使用
	model := client.GenerativeModel("gemini-2.5-pro-exp-03-25")

	// SafetySettingsを設定
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
	}

	// Function Callingの設定を追加
	model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingAuto,
		},
	}

	chat := model.StartChat()

	// システムプロンプトを設定
	systemContent := &genai.Content{
		Parts: []genai.Part{genai.Text(`You are an AI assistant that can use various tools to help users.
Please use the provided functions when necessary to accomplish tasks.
When a function is not needed, respond directly to the user.`)},
		Role: "model",
	}
	chat.History = append(chat.History, systemContent)

	return &GeminiClient{
		client: client,
		model:  model,
		chat:   chat,
		tools:  make(map[string]tools.Tool),
	}, nil
}

// RegisterTool はツールを登録します
func (g *GeminiClient) RegisterTool(tool tools.Tool) {
	g.tools[tool.Name()] = tool

	// ツールをFunctionDeclarationとして登録
	funcDecl := &genai.FunctionDeclaration{
		Name:        tool.Name(),
		Description: tool.Description(),
	}

	// ツールごとに適切なパラメータスキーマを設定
	switch tool.Name() {
	case "execCommand":
		funcDecl.Parameters = &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"command": {
					Type:        genai.TypeString,
					Description: "The shell command to execute (e.g. 'pwd', 'ls')",
				},
			},
			Required: []string{"command"},
		}
	case "complete":
		funcDecl.Parameters = &genai.Schema{
			Type:       genai.TypeObject,
			Properties: map[string]*genai.Schema{},
		}
	}

	// 既存のツールリストを取得
	var declarations []*genai.FunctionDeclaration
	if len(g.model.Tools) > 0 {
		declarations = g.model.Tools[0].FunctionDeclarations
	}

	// 新しいツールを追加
	declarations = append(declarations, funcDecl)

	// ツールを登録
	g.model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: declarations,
		},
	}

	// Function Callingの設定を更新
	g.model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingAny,
		},
	}
}

// ProcessMessage はユーザーのメッセージを処理し、適切な応答を返します。
// 必要に応じてツールを実行し、その結果を含めた応答を生成します。
func (g *GeminiClient) ProcessMessage(ctx context.Context, message string) (string, error) {
	// 空の入力をチェック
	if message == "" {
		return "", fmt.Errorf("empty input")
	}

	// 最初のプロンプト用の一時的な設定を作成
	tempConfig := g.model.ToolConfig
	tempTools := g.model.Tools

	// execCommandのみを使用可能に設定
	g.model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingAuto,
		},
	}
	g.model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "execCommand",
					Description: "シェルコマンドを実行するツールです",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"command": {
								Type:        genai.TypeString,
								Description: "The shell command to execute (e.g. 'pwd', 'ls')",
							},
						},
						Required: []string{"command"},
					},
				},
			},
		},
	}

	prompt := fmt.Sprintf(`
あなたは、シェルコマンドを実行できるAIアシスタントです。
ディレクトリの内容やパスを表示する要求には、execCommandを使用して適切なコマンドを実行してください。

例えば：
- 現在のディレクトリのパスを表示する場合: execCommand({"command": "pwd"})
- ディレクトリの内容を表示する場合: execCommand({"command": "ls"})

必ず日本語で応答してください。
コマンドの実行結果も日本語で説明してください。

ユーザーの要求: %s

適切なコマンドを実行して、結果を説明してください。
`, message)

	log.Printf("Sending prompt to LLM:\n%s\n", prompt)

	// メッセージを送信
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))

	// 設定を元に戻す
	g.model.ToolConfig = tempConfig
	g.model.Tools = tempTools

	if err != nil {
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	log.Printf("Received response from LLM:\n")
	logResponse(resp)

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no parts in response")
	}

	// レスポンスのパーツを確認
	for _, part := range candidate.Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			// テキストは無視して次のパーツを処理
			continue
		case genai.FunctionCall:
			tool, ok := g.tools[v.Name]
			if !ok {
				return "", fmt.Errorf("unknown tool: %s", v.Name)
			}

			log.Printf("Executing tool %s with args: %v\n", v.Name, v.Args)

			// ツールを実行
			result, err := tool.Execute(v.Args)
			if err != nil {
				return "", fmt.Errorf("tool execution failed: %v", err)
			}

			log.Printf("Tool execution result:\n%s\n", result)

			followUpPrompt := fmt.Sprintf(`
コマンド '%s' の実行結果を日本語で説明してください。
ファイルやディレクトリの情報を含めて、詳しく説明してください。

以下の形式で応答してください：
1. まず日本語で説明を書いてください（必須）
2. 説明の後に、completeツールを呼び出して完了を示してください（必須）

実行結果:
%s
`, v.Args["command"], result)

			log.Printf("Sending follow-up prompt to LLM:\n%s\n", followUpPrompt)

			// フォローアップ応答用の一時的な設定を作成
			tempConfig := g.model.ToolConfig
			g.model.ToolConfig = &genai.ToolConfig{
				FunctionCallingConfig: &genai.FunctionCallingConfig{
					Mode: genai.FunctionCallingAuto,
				},
			}
			g.model.Tools = []*genai.Tool{
				{
					FunctionDeclarations: []*genai.FunctionDeclaration{
						{
							Name:        "complete",
							Description: "タスクの完了を示すツールです。実行結果の説明を完了します。",
							Parameters:  &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{}},
						},
					},
				},
			}

			// ツールの実行結果を使用して新しい応答を生成
			followUpResp, err := g.model.GenerateContent(ctx, genai.Text(followUpPrompt))

			// 設定を元に戻す
			g.model.ToolConfig = tempConfig

			if err != nil {
				return "", fmt.Errorf("failed to get follow-up response: %v", err)
			}

			log.Printf("Received follow-up response from LLM:\n")
			logResponse(followUpResp)

			if len(followUpResp.Candidates) == 0 {
				return "", fmt.Errorf("no follow-up response from Gemini")
			}

			if len(followUpResp.Candidates[0].Content.Parts) == 0 {
				return "", fmt.Errorf("no parts in follow-up response")
			}

			var explanation string
			var completed bool

			// レスポンスのパーツを確認
			for _, part := range followUpResp.Candidates[0].Content.Parts {
				switch v := part.(type) {
				case genai.Text:
					explanation = string(v)
				case genai.FunctionCall:
					if v.Name == "complete" {
						tool, ok := g.tools[v.Name]
						if !ok {
							return "", fmt.Errorf("unknown tool: %s", v.Name)
						}
						_, err := tool.Execute(v.Args)
						if err != nil {
							return "", fmt.Errorf("complete tool execution failed: %v", err)
						}
						completed = true
					}
				}
			}

			if explanation == "" {
				return "", fmt.Errorf("no explanation found in response")
			}

			if !completed {
				return "", fmt.Errorf("task was not marked as complete")
			}

			return explanation, nil
		default:
			log.Printf("Unexpected part type: %T\n", part)
		}
	}

	// FunctionCallが見つからなかった場合はエラー
	return "", fmt.Errorf("no function call found in response")
}

// Close はクライアントを閉じます
func (g *GeminiClient) Close() {
	g.client.Close()
}
