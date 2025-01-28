package chat

import (
	"github.com/pankona/makasero/internal/models"
)

// APIClientAdapter は、api.Clientをchat.ChatClientインターフェースに適合させるアダプターです。
type APIClientAdapter struct {
	client interface {
		CreateChatCompletion(messages []models.ChatMessage) (string, error)
	}
}

// NewAPIClientAdapter は、新しいAPIClientAdapterを作成します。
func NewAPIClientAdapter(client interface {
	CreateChatCompletion(messages []models.ChatMessage) (string, error)
}) *APIClientAdapter {
	return &APIClientAdapter{client: client}
}

// CreateChatCompletion は、内部のAPIクライアントを使用してチャット完了を実行します。
func (a *APIClientAdapter) CreateChatCompletion(messages []Message) (string, error) {
	// メッセージ型の変換
	apiMessages := make([]models.ChatMessage, len(messages))
	for i, msg := range messages {
		apiMessages[i] = models.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return a.client.CreateChatCompletion(apiMessages)
}
