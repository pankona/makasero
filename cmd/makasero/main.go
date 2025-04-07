package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var (
	debug = flag.Bool("debug", false, "デバッグモードを有効にする")
)

func debugPrint(format string, args ...interface{}) {
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

func run() error {
	// コマンドライン引数の処理
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("コマンドを指定してください")
	}

	// APIキーの取得
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GOOGLE_API_KEY 環境変数が設定されていません")
	}

	// モデル名の取得（デフォルト: gemini-2.5-pro-exp-03-25）
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-2.5-pro-exp-03-25"
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

	// 関数定義の設定
	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "execCommand",
					Description: "ターミナルコマンドを実行し、その結果を返します",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"command": {
								Type:        genai.TypeString,
								Description: "実行するコマンド",
							},
							"args": {
								Type: genai.TypeArray,
								Items: &genai.Schema{
									Type: genai.TypeString,
								},
								Description: "コマンドの引数",
							},
						},
						Required: []string{"command"},
					},
				},
				{
					Name:        "complete",
					Description: "タスクが完了したことを示します",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"message": {
								Type:        genai.TypeString,
								Description: "完了メッセージ",
							},
						},
						Required: []string{"message"},
					},
				},
			},
		},
	}

	// チャットセッションの開始
	chat := model.StartChat()

	// ユーザーの入力を結合
	userInput := strings.Join(args, " ")

	// メッセージの送信と応答の取得
	resp, err := chat.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		return fmt.Errorf("メッセージの送信に失敗: %v", err)
	}

	shouldBreak := false
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
						debugPrint("Function call: %+v\n", p)
						if p.Name == "execCommand" {
							args, ok := p.Args["args"].([]interface{})
							if !ok {
								args = []interface{}{}
							}

							// コマンドの実行
							cmd := exec.Command(p.Args["command"].(string))
							for _, arg := range args {
								cmd.Args = append(cmd.Args, arg.(string))
							}
							fmt.Printf("実行するコマンド: %s %v\n", cmd.Path, cmd.Args)

							output, err := cmd.CombinedOutput()
							fmt.Printf("出力:\n%s\n", output)
							if err != nil {
								return fmt.Errorf("コマンドの実行に失敗: %v\n出力: %s", err, output)
							}

							// 実行結果を FunctionResponse として送信
							resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
								Name: "execCommand",
								Response: map[string]interface{}{
									"success": err == nil,
									"output":  string(output),
									"error":   err,
								},
							})
							if err != nil {
								return fmt.Errorf("実行結果の送信に失敗: %v", err)
							}

							// 続きのタスクを実行するために、ループを継続
							shouldBreak = false
						} else if p.Name == "complete" {
							// タスク完了の場合
							fmt.Printf("タスク完了: %s\n", p.Args["message"])
							return nil
						}
					case genai.Text:
						// テキスト応答の場合
						debugPrint("Text response: %s\n", p)
						fmt.Print(p)
					default:
						debugPrint("未知の応答タイプ: %T\n", part)
					}
				}
			}
		}
	}

	return nil
}
