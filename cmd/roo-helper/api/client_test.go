package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/cmd/roo-helper/models"
)

func TestNewClient(t *testing.T) {
	// 環境変数をテスト用に設定
	originalKey := os.Getenv("ROO_API_KEY")
	defer os.Setenv("ROO_API_KEY", originalKey)

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid api key",
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "missing api key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ROO_API_KEY", tt.apiKey)
			client, err := NewClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func setupMockServer(t *testing.T, statusCode int, response interface{}) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストの検証
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header with test-key")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header application/json")
		}

		// レスポンスの設定
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}))

	client := &Client{
		httpClient: server.Client(),
		apiKey:     "test-key",
		baseURL:    server.URL,
	}

	return server, client
}

func TestSendChatRequest(t *testing.T) {
	successResponse := models.ChatResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Choices: []struct {
			Message      models.ChatMessage `json:"message"`
			FinishReason string            `json:"finish_reason"`
		}{
			{
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: "Test response",
				},
				FinishReason: "stop",
			},
		},
	}

	errorResponse := models.ErrorResponse{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}{
			Message: "Test error",
			Type:    "invalid_request",
			Code:    "invalid_api_key",
		},
	}

	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		wantErr    bool
	}{
		{
			name:       "successful request",
			statusCode: http.StatusOK,
			response:   successResponse,
			wantErr:    false,
		},
		{
			name:       "api error",
			statusCode: http.StatusUnauthorized,
			response:   errorResponse,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client := setupMockServer(t, tt.statusCode, tt.response)
			defer server.Close()

			req := &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			}

			resp, err := client.SendChatRequest(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendChatRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp == nil {
					t.Error("SendChatRequest() returned nil response")
				} else if resp.Choices[0].Message.Content != "Test response" {
					t.Errorf("Expected response content 'Test response', got %s", resp.Choices[0].Message.Content)
				}
			}
		})
	}
}

func TestCreateChatCompletion(t *testing.T) {
	successResponse := models.ChatResponse{
		Choices: []struct {
			Message      models.ChatMessage `json:"message"`
			FinishReason string            `json:"finish_reason"`
		}{
			{
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: "Test completion",
				},
				FinishReason: "stop",
			},
		},
	}

	server, client := setupMockServer(t, http.StatusOK, successResponse)
	defer server.Close()

	tests := []struct {
		name     string
		messages []models.ChatMessage
		want     string
		wantErr  bool
	}{
		{
			name: "valid messages",
			messages: []models.ChatMessage{
				{Role: "user", Content: "Hello"},
			},
			want:    "Test completion",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.CreateChatCompletion(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateChatCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateChatCompletion() = %v, want %v", got, tt.want)
			}
		})
	}
}