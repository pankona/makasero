package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		wantErr bool
	}{
		{
			name:    "正常系：APIキーが設定されている",
			envKey:  "test-api-key",
			wantErr: false,
		},
		{
			name:    "異常系：APIキーが設定されていない",
			envKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OPENAI_API_KEY", tt.envKey)
			defer os.Unsetenv("OPENAI_API_KEY")

			client, err := NewClient()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.envKey, client.apiKey)
				assert.Equal(t, defaultBaseURL, client.baseURL)
			}
		})
	}
}

func TestSendChatRequest(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "正常系：APIリクエスト成功",
			statusCode: http.StatusOK,
			response: `{
				"id": "test-id",
				"object": "chat.completion",
				"created": 1234567890,
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "Hello!"
						},
						"finish_reason": "stop"
					}
				]
			}`,
			wantErr: false,
		},
		{
			name:       "異常系：APIエラー",
			statusCode: http.StatusBadRequest,
			response: `{
				"error": {
					"message": "Invalid request",
					"type": "invalid_request_error",
					"code": "invalid_api_key"
				}
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサーバーの設定
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// リクエストの検証
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			// テスト用クライアントの作成
			client := &Client{
				httpClient: server.Client(),
				apiKey:     "test-api-key",
				baseURL:    server.URL,
			}

			// テストリクエストの作成
			req := &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.ChatMessage{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			}

			// リクエストの実行
			resp, err := client.SendChatRequest(req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "Hello!", resp.Choices[0].Message.Content)
			}
		})
	}
}

func TestCreateChatCompletion(t *testing.T) {
	tests := []struct {
		name     string
		response *models.ChatResponse
		want     string
		wantErr  bool
	}{
		{
			name: "正常系：レスポンスあり",
			response: &models.ChatResponse{
				Choices: []struct {
					Message      models.ChatMessage `json:"message"`
					FinishReason string             `json:"finish_reason"`
				}{
					{
						Message: models.ChatMessage{
							Role:    "assistant",
							Content: "Hello!",
						},
					},
				},
			},
			want:    "Hello!",
			wantErr: false,
		},
		{
			name: "異常系：レスポンスなし",
			response: &models.ChatResponse{
				Choices: []struct {
					Message      models.ChatMessage `json:"message"`
					FinishReason string             `json:"finish_reason"`
				}{},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサーバーの設定
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response, _ := json.Marshal(tt.response)
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			}))
			defer server.Close()

			// テスト用クライアントの作成
			client := &Client{
				httpClient: server.Client(),
				apiKey:     "test-api-key",
				baseURL:    server.URL,
			}

			// テストの実行
			got, err := client.CreateChatCompletion([]models.ChatMessage{
				{
					Role:    "user",
					Content: "Hello",
				},
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
