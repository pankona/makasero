package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	chatCmd    = app.Command("chat", "AIとチャット")
	chatInput  = chatCmd.Arg("input", "チャット入力").Required().String()
	chatFile   = chatCmd.Flag("file", "対象ファイル").Short('f').String()
	chatApply  = chatCmd.Flag("apply", "変更を適用する").Short('a').Bool()
	chatBackup = chatCmd.Flag("backup-dir", "バックアップディレクトリ").Default("backups").String()
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
		result, err = executeCommand(client, "chat", *chatInput, *chatFile, *chatBackup)
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

	// システムメッセージの準備
	systemMessage := "あなたはコードレビューと改善提案を行う専門家です。"
	if targetFile != "" {
		systemMessage += "\n提案された変更は以下のフォーマットで返してください：\n\n" +
			"---PROPOSAL---\n" +
			"[変更の説明]\n\n" +
			"---CODE---\n" +
			"[変更後のコード全体]\n\n" +
			"---END---"
	}

	messages = append(messages, models.ChatMessage{
		Role:    "system",
		Content: systemMessage,
	})

	// ファイルが指定されている場合、その内容を読み取ってコンテキストとして追加
	if targetFile != "" {
		content, err := os.ReadFile(targetFile)
		if err != nil {
			return "", fmt.Errorf("ファイルの読み込みに失敗: %w", err)
		}

		fileContext := fmt.Sprintf("以下のファイルの内容について回答してください：\n\n```go\n%s\n```\n\n質問：%s",
			string(content), input)
		messages = append(messages, models.ChatMessage{
			Role:    "user",
			Content: fileContext,
		})
	} else {
		// JSON形式の場合はそのまま使用、そうでない場合は単純なメッセージとして扱う
		if err := json.Unmarshal([]byte(input), &messages); err != nil {
			messages = append(messages, models.ChatMessage{
				Role:    "user",
				Content: input,
			})
		}
	}

	response, err := client.CreateChatCompletion(messages)
	if err != nil {
		return "", err
	}

	// ファイルが指定され、かつ変更を適用する場合
	if targetFile != "" && *chatApply {
		proposal, code, err := parseAIResponse(response)
		if err != nil {
			return response, nil // パースに失敗しても元のレスポンスは返す
		}

		// バックアップの作成
		if err := createBackup(targetFile, backupDir); err != nil {
			return "", fmt.Errorf("バックアップの作成に失敗: %w", err)
		}

		// 変更の適用
		if err := applyChanges(targetFile, code); err != nil {
			return "", fmt.Errorf("変更の適用に失敗: %w", err)
		}

		return fmt.Sprintf("変更を適用しました。\n\n提案内容：\n%s", proposal), nil
	}

	return response, nil
}

// AIの応答から提案と変更後のコードを抽出
func parseAIResponse(response string) (proposal string, code string, err error) {
	proposalStart := strings.Index(response, "---PROPOSAL---")
	codeStart := strings.Index(response, "---CODE---")
	endMarker := strings.Index(response, "---END---")

	if proposalStart == -1 || codeStart == -1 || endMarker == -1 {
		return "", "", fmt.Errorf("不正なレスポンス形式")
	}

	proposal = strings.TrimSpace(response[proposalStart+len("---PROPOSAL---") : codeStart])
	code = strings.TrimSpace(response[codeStart+len("---CODE---") : endMarker])

	return proposal, code, nil
}

// バックアップの作成
func createBackup(filePath string, backupDir string) error {
	if backupDir == "" {
		return nil
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("バックアップディレクトリの作成に失敗: %w", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗: %w", err)
	}

	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%d.bak",
		filepath.Base(filePath), time.Now().Unix()))

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("バックアップファイルの作成に失敗: %w", err)
	}

	return nil
}

// 変更の適用
func applyChanges(filePath string, newContent string) error {
	return os.WriteFile(filePath, []byte(newContent), 0644)
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
