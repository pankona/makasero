package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/pankona/makasero/tools"
	"google.golang.org/api/option"
)

// GeminiClient はGeminiとの通信を管理するクライアントです
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
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
	return &GeminiClient{
		client: client,
		model:  model,
		tools:  make(map[string]tools.Tool),
	}, nil
}

// RegisterTool はツールを登録します
func (g *GeminiClient) RegisterTool(tool tools.Tool) {
	g.tools[tool.Name()] = tool
}

// GenerateContent はGeminiにプロンプトを送信し、レスポンスを取得します
func (g *GeminiClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	// ツールの説明を生成
	toolsDesc := make([]string, 0, len(g.tools))
	for _, tool := range g.tools {
		toolsDesc = append(toolsDesc, fmt.Sprintf("- %s: %s", tool.Name(), tool.Description()))
	}

	// システムプロンプトを設定
	systemPrompt := fmt.Sprintf(`You are an AI assistant with access to the following tools:

%s

When you need to use a tool, respond with a JSON object in the following format:
{
    "tool": "tool_name",
    "args": {
        "arg1": "value1",
        "arg2": "value2"
    }
}

Otherwise, respond with a direct answer.`, strings.Join(toolsDesc, "\n"))

	// プロンプトを結合
	fullPrompt := fmt.Sprintf("%s\n\nUser: %s", systemPrompt, prompt)

	resp, err := g.model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

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
