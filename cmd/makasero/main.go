package main

import (
	"bufio"
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
	"github.com/pankona/makasero/internal/proposal"
)

var (
	app = kingpin.New("makasero", "コード改善支援CLIツール")

	// explainコマンド
	explainCmd  = app.Command("explain", "コードの説明を生成")
	explainCode = explainCmd.Arg("code", "説明するコードまたはファイルパス").Required().String()

	// chatコマンド
	chatCmd     = app.Command("chat", "AIとチャット")
	chatInput   = chatCmd.Arg("input", "チャット入力").Required().String()
	chatFile    = chatCmd.Flag("file", "対象ファイル").Short('f').String()
	chatApply   = chatCmd.Flag("apply", "変更を適用する").Short('a').Bool()
	chatBackup  = chatCmd.Flag("backup-dir", "バックアップディレクトリ").Default("backups").String()
	chatAutoYes = chatCmd.Flag("yes", "確認をスキップして自動的に承認").Short('y').Bool()
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
	// コードがファイルパスのような形式かチェック
	if strings.Contains(code, "/") || strings.Contains(code, "\\") {
		// ファイルパスとして処理
		info, err := os.Stat(code)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("ファイルが存在しません: %s", code)
			}
			return "", fmt.Errorf("ファイルの状態確認に失敗: %w", err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("指定されたパスはディレクトリです: %s", code)
		}

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

	response, err := client.CreateChatCompletion(messages)
	if err != nil {
		return "", fmt.Errorf("APIリクエストに失敗: %w", err)
	}

	return response, nil
}

func executeChat(client api.APIClient, input string, targetFile string, backupDir string) (string, error) {
	// ファイルが指定されている場合、その内容を読み込む
	var fileContent string
	if targetFile != "" {
		content, err := os.ReadFile(targetFile)
		if err != nil {
			return "", fmt.Errorf("ファイルの読み込みに失敗: %w", err)
		}
		fileContent = string(content)
		input = fmt.Sprintf("以下のコードを改善してください:\n\n%s\n\n%s", fileContent, input)
	}

	// APIにリクエストを送信
	messages := []models.ChatMessage{
		{
			Role: "system",
			Content: "あなたはコードレビューと改善提案を行う専門家です。\n" +
				"提案された変更は以下のフォーマットで返してください：\n\n" +
				"---PROPOSAL---\n" +
				"[変更の説明]\n\n" +
				"---CODE---\n" +
				"[変更後のコード全体]\n\n" +
				"---END---",
		},
		{
			Role:    "user",
			Content: input,
		},
	}

	response, err := client.CreateChatCompletion(messages)
	if err != nil {
		return "", fmt.Errorf("APIリクエストに失敗: %w", err)
	}

	// ファイルが指定されている場合、変更を処理
	if targetFile != "" {
		proposal, code, err := parseAIResponse(response)
		if err != nil {
			return response, nil // パースに失敗しても元のレスポンスは返す
		}

		// 差分を表示し、ユーザーの承認を得る
		approved, err := promptForApproval(fileContent, code)
		if err != nil {
			return "", fmt.Errorf("ユーザー承認の取得に失敗: %w", err)
		}

		if !approved {
			return "変更は取り消されました。", nil
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

// 変更の差分を表示し、ユーザーの承認を得る
func promptForApproval(originalCode, proposedCode string) (bool, error) {
	diffUtil := proposal.NewDiffUtility()
	diff := diffUtil.FormatDiff(originalCode, proposedCode)

	fmt.Println("\n変更内容:")
	fmt.Println(diff)

	// テスト時または自動承認オプションが指定されている場合は自動的に承認
	if os.Getenv("MAKASERO_TEST") == "1" || *chatAutoYes {
		return true, nil
	}

	fmt.Print("\n変更を適用しますか？ [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("入力の読み取りに失敗: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y", nil
}

// バックアップの作成
func createBackup(filePath, backupDir string) error {
	if backupDir == "" {
		backupDir = "backups"
	}

	// バックアップディレクトリの作成
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("バックアップディレクトリの作成に失敗: %w", err)
	}

	backupPath := filepath.Join(backupDir,
		fmt.Sprintf("%s.%d.bak",
			filepath.Base(filePath),
			time.Now().UnixNano()))

	// ファイルをコピー
	input, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗: %w", err)
	}

	if err := os.WriteFile(backupPath, input, 0644); err != nil {
		return fmt.Errorf("バックアップの書き込みに失敗: %w", err)
	}

	return nil
}

// 変更の適用
func applyChanges(filePath, newContent string) error {
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗: %w", err)
	}
	return nil
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
