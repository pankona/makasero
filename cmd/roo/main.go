package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/api"
	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/models"
)

// APIClient インターフェースを定義
type APIClient interface {
	CreateChatCompletion(messages []models.ChatMessage) (string, error)
}

func main() {
	// コマンドライン引数の設定
	command := flag.String("command", "", "Command to execute (required)")
	input := flag.String("input", "", "Input data")
	flag.Parse()

	// コマンドの必須チェック
	if *command == "" {
		log.Fatal("command is required")
	}

	// APIクライアントの初期化
	client, err := api.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize API client: %v", err)
	}

	// コマンドの実行
	result, err := executeCommand(client, *command, *input)
	if err != nil {
		response := models.Response{
			Success: false,
			Error:   err.Error(),
		}
		outputResponse(response)
		os.Exit(1)
	}

	// 成功レスポンスの出力
	response := models.Response{
		Success: true,
		Data:    result,
	}
	outputResponse(response)
}

// executeCommand executes the specified command with the given input
func executeCommand(client APIClient, command, input string) (interface{}, error) {
	switch command {
	case "explain":
		return executeExplain(client, input)
	case "chat":
		return executeChat(client, input)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// executeExplain handles the explain command
func executeExplain(client APIClient, code string) (string, error) {
	messages := []models.ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant that explains code.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Please explain this code:\n\n%s", code),
		},
	}

	return client.CreateChatCompletion(messages)
}

// executeChat handles the chat command
func executeChat(client APIClient, input string) (string, error) {
	var messages []models.ChatMessage
	if err := json.Unmarshal([]byte(input), &messages); err != nil {
		return "", fmt.Errorf("failed to parse chat messages: %w", err)
	}

	return client.CreateChatCompletion(messages)
}

// outputResponse outputs the response as JSON to stdout
func outputResponse(response models.Response) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		log.Fatalf("Failed to encode response: %v", err)
	}
}
