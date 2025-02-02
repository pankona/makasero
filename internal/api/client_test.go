package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pankona/makasero/internal/models"
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
			}
		})
	}
}

func TestCreateChatCompletion(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
		want       string
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
			want:    "Hello!",
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
			want:    "",
		},
		{
			name:       "異常系：レスポンスなし",
			statusCode: http.StatusOK,
			response: `{
				"id": "test-id",
				"object": "chat.completion",
				"created": 1234567890,
				"choices": []
			}`,
			wantErr: true,
			want:    "",
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

func TestMockClient(t *testing.T) {
	tests := []struct {
		name     string
		response string
		err      error
		wantErr  bool
	}{
		{
			name:     "正常系：レスポンスあり",
			response: "Hello!",
			err:      nil,
			wantErr:  false,
		},
		{
			name:     "異常系：エラーあり",
			response: "",
			err:      assert.AnError,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockClient{
				Response: tt.response,
				Err:      tt.err,
			}

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
				assert.Equal(t, tt.response, got)
			}
		})
	}
}

func TestSetMockClient(t *testing.T) {
	// 初期状態の確認
	assert.Nil(t, mockAPIClient)
	assert.Nil(t, mockAPIErr)

	// モッククライアントの設定
	mockClient := &MockClient{
		Response: "Hello!",
		Err:      nil,
	}
	SetMockClient(mockClient, assert.AnError)

	// 設定後の状態確認
	assert.Equal(t, mockClient, mockAPIClient)
	assert.Equal(t, assert.AnError, mockAPIErr)

	// モッククライアントのリセット
	SetMockClient(nil, nil)
	assert.Nil(t, mockAPIClient)
	assert.Nil(t, mockAPIErr)
}
