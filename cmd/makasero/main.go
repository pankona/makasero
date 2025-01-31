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
	chatCmd    = app.Command("chat", "AIとチャット")
	chatFile   = chatCmd.Flag("file", "対象ファイル").Short('f').String()
	chatBackup = chatCmd.Flag("backup-dir", "バックアップディレクトリ").Default("backups").String()
	chatResume = chatCmd.Flag("resume", "過去のセッションを再開").Short('r').String()
	chatList   = chatCmd.Flag("list", "過去のセッション一覧を表示").Short('l').Bool()
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
		result, err = executeChat(client, *chatFile, *chatBackup)
	}

	if err != nil {
		outputResponse(models.Response{
			Success: false,
			Error:   err.Error(),
		})
		os.Exit(1)
	}

	if command != chatCmd.FullCommand() {
		outputResponse(models.Response{
			Success: true,
			Data:    result,
		})
	}
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

func executeChat(client api.APIClient, targetFile string, backupDir string) (string, error) {
	var messages []models.ChatMessage
	var fileContent string
	var session *models.ChatSession

	// セッションディレクトリの設定
	sessionDir := filepath.Join(os.Getenv("HOME"), ".makasero", "sessions")

	// セッション一覧の表示
	if *chatList {
		sessions, err := models.ListSessions(sessionDir)
		if err != nil {
			return "", fmt.Errorf("セッション一覧の取得に失敗: %w", err)
		}

		if len(sessions) == 0 {
			return "保存されたセッションはありません。", nil
		}

		fmt.Println("保存されたセッション一覧:")
		for _, s := range sessions {
			fmt.Printf("ID: %s\n", s.ID)
			fmt.Printf("開始時刻: %s\n", s.StartTime.Format("2006-01-02 15:04:05"))
			if s.FilePath != "" {
				fmt.Printf("ファイル: %s\n", s.FilePath)
			}

			// 最新のメッセージを表示（最大3つ）
			fmt.Println("最新の会話:")
			messageCount := len(s.Messages)
			start := messageCount - 6 // 3往復分（ユーザー＋AI）
			if start < 0 {
				start = 0
			}
			for i := start; i < messageCount; i++ {
				msg := s.Messages[i]
				if msg.Role == "system" {
					continue
				}
				prefix := "  AI> "
				if msg.Role == "user" {
					prefix = "  You> "
				}
				// メッセージを1行に省略
				content := msg.Content
				if len(content) > 60 {
					content = content[:57] + "..."
				}
				fmt.Printf("%s%s\n", prefix, content)
			}
			fmt.Println("---")
		}
		return "", nil
	}

	// セッションの再開
	if *chatResume != "" {
		var err error
		session, err = models.LoadSession(sessionDir, *chatResume)
		if err != nil {
			return "", fmt.Errorf("セッションの読み込みに失敗: %w", err)
		}
		messages = session.Messages
		targetFile = session.FilePath
		fmt.Printf("セッション '%s' を再開します。\n", session.ID)
	} else {
		// 新しいセッションの作成
		session = &models.ChatSession{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			StartTime: time.Now(),
			FilePath:  targetFile,
		}
	}

	// システムメッセージの設定
	if len(messages) == 0 {
		messages = append(messages, models.ChatMessage{
			Role: "system",
			Content: "あなたはコードレビューと改善提案を行う専門家です。\n" +
				"コードの変更を提案する場合は以下のフォーマットで返してください：\n\n" +
				"---PROPOSAL---\n" +
				"[変更の説明]\n\n" +
				"---CODE---\n" +
				"[変更後のコード全体]\n\n" +
				"---END---",
		})
	}

	// ファイルが指定されている場合、その内容を読み込む
	if targetFile != "" {
		content, err := os.ReadFile(targetFile)
		if err != nil {
			return "", fmt.Errorf("ファイルの読み込みに失敗: %w", err)
		}
		fileContent = string(content)
		if len(messages) == 1 { // システムメッセージのみの場合
			messages = append(messages, models.ChatMessage{
				Role:    "system",
				Content: fmt.Sprintf("以下のファイルについて会話を行います：\n\n```go\n%s\n```", fileContent),
			})
		}
	}

	fmt.Println("チャットを開始します。終了するには 'exit' または 'quit' と入力してください。")
	if targetFile != "" {
		fmt.Printf("モード: ファイル編集 (%s)\n", targetFile)
	}
	fmt.Printf("セッションID: %s\n", session.ID)
	reader := bufio.NewReader(os.Stdin)

	for {
		if targetFile != "" {
			fmt.Print("\nmakasero (file)> ")
		} else {
			fmt.Print("\nmakasero> ")
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("入力の読み取りに失敗: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 終了コマンドの確認
		if input == "exit" || input == "quit" {
			// セッションの保存
			session.Messages = messages
			if err := session.SaveSession(sessionDir); err != nil {
				fmt.Printf("警告: セッションの保存に失敗しました: %v\n", err)
			} else {
				fmt.Printf("セッションを保存しました。再開するには: makasero chat -r %s\n", session.ID)
			}
			fmt.Println("チャットを終了します。")
			return "", nil
		}

		// ユーザーの入力をメッセージに追加
		messages = append(messages, models.ChatMessage{
			Role:    "user",
			Content: input,
		})

		// APIにリクエストを送信
		response, err := client.CreateChatCompletion(messages)
		if err != nil {
			return "", fmt.Errorf("APIリクエストに失敗: %w", err)
		}

		// AIの応答をメッセージに追加
		messages = append(messages, models.ChatMessage{
			Role:    "assistant",
			Content: response,
		})

		// ファイルが指定されている場合、変更提案を処理
		if targetFile != "" && strings.Contains(response, "---PROPOSAL---") {
			_, code, err := parseAIResponse(response)
			if err == nil {
				// 差分を表示し、ユーザーの承認を得る
				approved, err := promptForApproval(fileContent, code)
				if err != nil {
					fmt.Printf("エラー: %v\n", err)
					continue
				}

				if approved {
					// バックアップの作成
					if err := createBackup(targetFile, backupDir); err != nil {
						fmt.Printf("バックアップの作成に失敗: %v\n", err)
						continue
					}

					// 変更の適用
					if err := applyChanges(targetFile, code); err != nil {
						fmt.Printf("変更の適用に失敗: %v\n", err)
						continue
					}

					fileContent = code // 更新されたコードを保持
					fmt.Println("変更を適用しました。")
				} else {
					fmt.Println("変更は取り消されました。")
				}
			}
		}

		// AIの応答を表示
		fmt.Println("\n" + response)

		// セッションの自動保存
		session.Messages = messages
		if err := session.SaveSession(sessionDir); err != nil {
			fmt.Printf("警告: セッションの自動保存に失敗しました: %v\n", err)
		}
	}
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
	if os.Getenv("MAKASERO_TEST") == "1" {
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
		return executeChat(client, targetFile, backupDir)
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
