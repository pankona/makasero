package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var (
	debug      = flag.Bool("debug", false, "デバッグモード")
	promptFile = flag.String("f", "", "プロンプトファイル")
)

func debugPrint(format string, args ...any) {
	if *debug {
		fmt.Printf("[DEBUG] "+format, args...)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func readPromptFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("プロンプトファイルの読み込みに失敗: %v", err)
	}
	return string(content), nil
}

func run() error {
	// コマンドライン引数の処理
	flag.Parse()

	// セッション一覧表示の処理
	if *listSessionsFlag {
		return listSessions()
	}

	// 会話履歴全文表示の処理
	if *showHistory != "" {
		return showSessionHistory(*showHistory)
	}

	args := flag.Args()

	// プロンプトの取得
	var userInput string
	if *promptFile != "" {
		// ファイルからプロンプトを読み込む
		prompt, err := readPromptFromFile(*promptFile)
		if err != nil {
			return err
		}
		userInput = prompt
		fmt.Printf("プロンプトファイルから読み込んだ内容:\n%s\n", userInput)
	} else if len(args) > 0 {
		// コマンドライン引数からプロンプトを取得
		userInput = strings.Join(args, " ")
		fmt.Printf("コマンドライン引数から取得したプロンプト:\n%s\n", userInput)
	} else {
		return fmt.Errorf("プロンプトを指定してください（コマンドライン引数または -f オプション）")
	}

	// APIキーの取得
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY 環境変数が設定されていません")
	}

	// モデル名の取得（デフォルト: gemini-2.0-flash-lite）
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-2.0-flash-lite"
	}

	// コンテキストの作成
	ctx := context.Background()

	// クライアントの初期化
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("クライアントの初期化に失敗: %v", err)
	}
	defer client.Close()

	// モデルの初期化
	model := client.GenerativeModel(modelName)

	// 関数定義から FunctionDeclaration のスライスを作成
	var declarations []*genai.FunctionDeclaration
	for _, fn := range functions {
		declarations = append(declarations, fn.Declaration)
	}

	// モデルに設定
	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: declarations,
		},
	}

	// セッションの読み込み
	var session *Session
	if *sessionID != "" {
		var err error
		session, err = loadSession(*sessionID)
		if err != nil {
			return err
		}
	} else {
		// 新規セッション
		session = &Session{
			ID:        generateSessionID(),
			CreatedAt: time.Now(),
		}
	}

	// チャットセッションを開始
	chat := model.StartChat()
	if len(session.History) > 0 {
		chat.History = session.History
	}

	// メッセージの送信と応答の取得
	fmt.Printf("\nAIに送信するメッセージ:\n%s\n\n", userInput)
	resp, err := chat.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		// エラーが発生しても、それまでの履歴は保存
		session.History = chat.History
		saveSession(session)
		return fmt.Errorf("メッセージの送信に失敗: %v", err)
	}

	var shouldBreak bool
	for !shouldBreak {
		shouldBreak = true

		// レスポンスの処理
		if len(resp.Candidates) > 0 {
			cand := resp.Candidates[0]
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					switch p := part.(type) {
					case genai.FunctionCall:
						// 関数呼び出しの場合
						fmt.Printf("\n関数呼び出し: %s\n", p.Name)
						fmt.Printf("引数: %+v\n", p.Args)

						if p.Name == "complete" || p.Name == "askQuestion" {
							session.History = chat.History
							session.UpdatedAt = time.Now()
							if err := saveSession(session); err != nil {
								return err
							}
							fmt.Printf("\nセッションID: %s\n", session.ID)
							return nil
						}

						fn, exists := functions[p.Name]
						if !exists {
							return fmt.Errorf("unknown function: %s", p.Name)
						}

						result, err := fn.Handler(ctx, p.Args)
						if err != nil {
							return fmt.Errorf("function %s failed: %v", p.Name, err)
						}

						// 実行結果を FunctionResponse として送信
						resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
							Name:     p.Name,
							Response: result,
						})
						if err != nil {
							return fmt.Errorf("実行結果の送信に失敗: %v", err)
						}

						// complete 関数以外の場合は続きのタスクを実行するために、ループを継続
						shouldBreak = false
					case genai.Text:
						// テキスト応答の場合
						fmt.Printf("\nAIからの応答:\n%s\n", p)
					default:
						debugPrint("未知の応答タイプ: %T\n", part)
					}
				}
			}
		}
	}

	return nil
}
