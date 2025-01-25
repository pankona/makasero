package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/cmd/roo-helper/models"
)

const (
	defaultTimeout = 30 * time.Second
	defaultBaseURL = "https://api.openai.com/v1"
)

// Client represents an API client for making requests
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// NewClient creates a new API client instance
func NewClient() (*Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
	}, nil
}

// SendChatRequest sends a chat completion request to the API
func (c *Client) SendChatRequest(req *models.ChatRequest) (*models.ChatResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	request, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", response.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
	}

	var chatResp models.ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// CreateChatCompletion is a helper function to create a chat completion
func (c *Client) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
	req := &models.ChatRequest{
		Model:    "gpt-4",
		Messages: messages,
	}

	resp, err := c.SendChatRequest(req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices available")
	}

	return resp.Choices[0].Message.Content, nil
}
