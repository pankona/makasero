package chat

import (
	"errors"
	"testing"

	"github.com/pankona/makasero/internal/models"
	"github.com/stretchr/testify/assert"
)

// モックAPIクライアント
type mockAPIClient struct {
	response string
	err      error
}

func (m *mockAPIClient) CreateChatCompletion(messages []models.ChatMessage) (string, error) {
	return m.response, m.err
}

func TestNewAPIClientAdapter(t *testing.T) {
	tests := []struct {
		name   string
		client *mockAPIClient
	}{
		{
			name: "正常系：アダプターの作成",
			client: &mockAPIClient{
				response: "テスト応答",
				err:      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAPIClientAdapter(tt.client)
			assert.NotNil(t, adapter)
			assert.Equal(t, tt.client, adapter.client)
		})
	}
}

func TestAPIClientAdapter_CreateChatCompletion(t *testing.T) {
	tests := []struct {
		name     string
		client   *mockAPIClient
		messages []Message
		want     string
		wantErr  bool
	}{
		{
			name: "正常系：メッセージ変換と応答",
			client: &mockAPIClient{
				response: "テスト応答",
				err:      nil,
			},
			messages: []Message{
				{Role: "user", Content: "こんにちは"},
				{Role: "assistant", Content: "はい、こんにちは"},
			},
			want:    "テスト応答",
			wantErr: false,
		},
		{
			name: "異常系：APIエラー",
			client: &mockAPIClient{
				response: "",
				err:      errors.New("APIエラー"),
			},
			messages: []Message{
				{Role: "user", Content: "こんにちは"},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "正常系：空のメッセージリスト",
			client: &mockAPIClient{
				response: "空の応答",
				err:      nil,
			},
			messages: []Message{},
			want:     "空の応答",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAPIClientAdapter(tt.client)
			got, err := adapter.CreateChatCompletion(tt.messages)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
