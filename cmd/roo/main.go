package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/api"
	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/chat"
	"github.com/rooveterinaryinc/hello-vim-plugin-2/internal/models"
)

// APIClient はAPIとの対話を行うインターフェースです。
type APIClient interface {
	CreateChatCompletion(messages []models.ChatMessage) (string, error)
}

func main() {
	// コマンドライン引数の設定
	var (
		input     = flag.String("input", "", "Input text or file path")
		backupDir = flag.String("backup-dir", "", "Directory for backup files")
	)
	flag.Parse()

	// 位置引数からコマンドを取得
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("command is required (chat or explain)")
	}
	command := args[0]

	// 入力テキストの取得（標準入力、位置引数、または-input）
	inputText := *input
	if inputText == "" && len(args) > 1 {
		inputText = args[1]
	}

	// 標準入力があれば、それを読み取る
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		reader := bufio.NewReader(os.Stdin)
		var sb strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("標準入力の読み取りに失敗: %v", err)
			}
			sb.WriteString(line)
		}
		if sb.Len() > 0 {
			if inputText != "" {
				inputText = fmt.Sprintf("%s\n\n%s", inputText, sb.String())
			} else {
				inputText = sb.String()
			}
		}
	}

	// APIクライアントの初期化
	client, err := api.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize API client: %v", err)
	}

	// コマンドの実行
	result, err := executeCommand(client, command, inputText, *backupDir)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(result)
}

// executeCommand executes the specified command with the given input
func executeCommand(client APIClient, command, input, backupDir string) (string, error) {
	switch command {
	case "explain":
		return executeExplain(client, input)
	case "chat":
		return executeChat(client, input, backupDir)
	default:
		return "", fmt.Errorf("unknown command: %s", command)
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
func executeChat(client APIClient, input, backupDir string) (string, error) {
	// バックアップディレクトリの設定
	if backupDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		backupDir = filepath.Join(homeDir, ".roo", "backups")
	}

	// アダプターを使用してチャット実行器を初期化
	adapter := chat.NewAPIClientAdapter(client)
	executor, err := chat.NewExecutor(adapter, backupDir)
	if err != nil {
		return "", fmt.Errorf("failed to initialize chat executor: %w", err)
	}

	// チャットの実行
	return executor.Execute(input)
}
