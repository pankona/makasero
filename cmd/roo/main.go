package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/api"
	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/models"
	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/proposal"
)

// APIClient インターフェースを定義
type APIClient interface {
	CreateChatCompletion(messages []models.ChatMessage) (string, error)
}

func main() {
	// コマンドライン引数の設定
	command := flag.String("command", "", "Command to execute (required)")
	input := flag.String("input", "", "Input data")
	mode := flag.String("mode", "patch", "Proposal mode (patch or full)")
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
	result, err := executeCommand(client, *command, *input, *mode)
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
func executeCommand(client APIClient, command, input, mode string) (interface{}, error) {
	switch command {
	case "explain":
		return executeExplain(client, input)
	case "chat":
		return executeChat(client, input)
	case "propose":
		// 型アサーションを使用してapi.Clientを取得
		apiClient, ok := client.(*api.Client)
		if !ok {
			return nil, fmt.Errorf("invalid client type for propose command")
		}
		return executePropose(apiClient, input, mode)
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

// executePropose handles the propose command
func executePropose(client *api.Client, filePath, mode string) (interface{}, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path is required for propose command")
	}

	// モードの検証
	var applyMode proposal.ApplyMode
	switch mode {
	case "patch":
		applyMode = proposal.ApplyModePatch
	case "full":
		applyMode = proposal.ApplyModeFull
	default:
		return nil, fmt.Errorf("invalid mode: %s (must be 'patch' or 'full')", mode)
	}

	// ProposalManagerの初期化
	manager := proposal.NewManager(client, proposal.ConsoleUserApprover())

	// 提案の生成
	prop, err := manager.GenerateProposal(filePath, applyMode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proposal: %w", err)
	}

	// 提案の適用
	if err := manager.ApplyProposal(prop); err != nil {
		return nil, fmt.Errorf("failed to apply proposal: %w", err)
	}

	return map[string]interface{}{
		"message": "コードの変更が正常に適用されました",
		"file":    filePath,
		"mode":    mode,
	}, nil
}

// outputResponse outputs the response as JSON to stdout
func outputResponse(response models.Response) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		log.Fatalf("Failed to encode response: %v", err)
	}
}
