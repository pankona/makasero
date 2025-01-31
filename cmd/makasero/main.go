package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/pankona/makasero/internal/api"
	"github.com/pankona/makasero/internal/models"
)

var (
	app = kingpin.New("makasero", "コード改善支援CLIツール")

	// explainコマンド
	explainCmd  = app.Command("explain", "コードの説明を生成")
	explainCode = explainCmd.Arg("code", "説明するコードまたはファイルパス").Required().String()

	// chatコマンド
	chatCmd   = app.Command("chat", "AIとチャット")
	chatInput = chatCmd.Arg("input", "チャット入力").Required().String()
)

func main() {
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	// APIクライアントの初期化
	client, err := api.NewClient()
	if err != nil {
		log.Fatalf("APIクライアントの初期化に失敗: %v", err)
	}

	// コマンドの実行
	var result string
	switch command {
	case explainCmd.FullCommand():
		result, err = executeCommand(client, "explain", *explainCode, "", "")
	case chatCmd.FullCommand():
		result, err = executeCommand(client, "chat", *chatInput, "", "")
	}

	if err != nil {
		outputResponse(models.Response{
			Success: false,
			Error:   err.Error(),
		})
		os.Exit(1)
	}

	outputResponse(models.Response{
		Success: true,
		Data:    result,
	})
}

func executeExplain(client api.APIClient, code string) (string, error) {
	// ファイルが存在する場合はファイルから読み取る
	if _, err := os.Stat(code); err == nil {
		content, err := os.ReadFile(code)
		if err != nil {
			return "", fmt.Errorf("ファイルの読み込みに失敗: %w", err)
		}
		code = string(content)
	}

	messages := []models.ChatMessage{
		{
			Role:    "system",
			Content: "あなたはコードの説明を生成する専門家です。",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("以下のコードを説明してください：\n\n```go\n%s\n```", code),
		},
	}

	return client.CreateChatCompletion(messages)
}

func executeChat(client api.APIClient, input string, targetFile string, backupDir string) (string, error) {
	var messages []models.ChatMessage
	if err := json.Unmarshal([]byte(input), &messages); err != nil {
		// JSON形式でない場合は、単純なメッセージとして扱う
		messages = []models.ChatMessage{
			{
				Role:    "user",
				Content: input,
			},
		}
	}

	return client.CreateChatCompletion(messages)
}

func executeCommand(client api.APIClient, command, input, targetFile, backupDir string) (string, error) {
	switch command {
	case "explain":
		return executeExplain(client, input)
	case "chat":
		return executeChat(client, input, targetFile, backupDir)
	default:
		return "", fmt.Errorf("不明なコマンド: %s", command)
	}
}

func outputResponse(response models.Response) {
	json, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("JSONの生成に失敗: %v", err)
	}
	fmt.Println(string(json))
}
