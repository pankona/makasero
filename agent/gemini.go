package agent

import (
	"context"
	"fmt"

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

// NewGeminiClient は新しいGeminiClientを作成します
func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel("gemini-pro")

	// ツールの設定
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
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"command": {
					Type:        genai.TypeString,
					Description: "The command to execute",
				},
			},
			Required: []string{"command"},
		},
	}

	// ツールを登録
	g.model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		},
	}
}

// ProcessMessage はユーザーのメッセージを処理し、適切な応答を返します。
// 必要に応じてツールを実行し、その結果を含めた応答を生成します。
func (g *GeminiClient) ProcessMessage(ctx context.Context, message string) (string, error) {
	// ユーザーの入力をチャット履歴に追加
	userContent := &genai.Content{
		Parts: []genai.Part{genai.Text(message)},
		Role:  "user",
	}

	// メッセージを送信
	resp, err := g.chat.SendMessage(ctx, userContent.Parts...)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	// レスポンスのパーツを確認
	for _, part := range resp.Candidates[0].Content.Parts {
		// FunctionCallのチェック
		if funcCall, ok := part.(*genai.FunctionCall); ok {
			tool, ok := g.tools[funcCall.Name]
			if !ok {
				return "", fmt.Errorf("unknown tool: %s", funcCall.Name)
			}

			// ツールを実行
			result, err := tool.Execute(funcCall.Args)
			if err != nil {
				return "", fmt.Errorf("tool execution failed: %v", err)
			}

			// ツールの実行結果をチャット履歴に追加
			resultContent := &genai.Content{
				Parts: []genai.Part{genai.Text(fmt.Sprintf("Function '%s' returned: %s", funcCall.Name, result))},
				Role:  "function",
			}
			g.chat.History = append(g.chat.History, resultContent)

			// フォローアップの応答を取得
			followUpResp, err := g.chat.SendMessage(ctx, genai.Text("Please provide a final answer based on the function result."))
			if err != nil {
				return "", fmt.Errorf("failed to get follow-up response: %v", err)
			}

			if len(followUpResp.Candidates) == 0 {
				return "", fmt.Errorf("no follow-up response from Gemini")
			}

			text, ok := followUpResp.Candidates[0].Content.Parts[0].(genai.Text)
			if !ok {
				return "", fmt.Errorf("unexpected response type")
			}

			return string(text), nil
		}
	}

	// 通常の応答を返す
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response type")
	}

	return string(text), nil
}

// Close はクライアントを閉じます
func (g *GeminiClient) Close() {
	g.client.Close()
}
