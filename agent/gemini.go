package agent

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient はGemini APIを使用するLLMClientの実装です
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiClient は新しいGeminiClientを作成します
func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	model := client.GenerativeModel("gemini-pro")

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// GenerateContent はGemini APIを使用してテキストを生成します
func (c *GeminiClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	content := resp.Candidates[0].Content
	if len(content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in response")
	}

	part := content.Parts[0]
	if part == nil {
		return "", fmt.Errorf("nil part in response")
	}

	textPart, ok := part.(genai.Text)
	if !ok {
		return "", fmt.Errorf("part is not text: %T", part)
	}

	return string(textPart), nil
}

// Close はクライアントのリソースを解放します
func (c *GeminiClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
