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
	debug      = flag.Bool("debug", false, "デバッグモードを有効にする")
	promptFile = flag.String("f", "", "プロンプトファイルのパス")
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
					Name:        "getGitHubIssue",
					Description: "GitHubのIssueの詳細を取得します",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"repository": {
								Type:        genai.TypeString,
								Description: "リポジトリ名（例: owner/repo）",
							},
							"issueNumber": {
								Type:        genai.TypeInteger,
								Description: "Issue番号",
							},
						},
						Required: []string{"repository", "issueNumber"},
					},
				},
				{
					Name:        "gitStatus",
					Description: "Gitのステータスを取得します",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {
								Type:        genai.TypeString,
								Description: "確認するパス（オプション）",
							},
						},
					},
				},
				{
					Name:        "gitAdd",
					Description: "指定されたファイルをGitのステージングエリアに追加します",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"paths": {
								Type: genai.TypeArray,
								Items: &genai.Schema{
									Type: genai.TypeString,
								},
								Description: "追加するファイルのパス（配列）",
							},
							"all": {
								Type:        genai.TypeBoolean,
								Description: "すべての変更を追加するかどうか",
							},
						},
					},
				},
				{
					Name:        "gitCommit",
					Description: "ステージングされた変更をコミットします",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"message": {
								Type:        genai.TypeString,
								Description: "コミットメッセージ",
							},
							"type": {
								Type:        genai.TypeString,
								Description: "コミットの種類（feat, fix, docs, style, refactor, test, chore など）",
							},
							"scope": {
								Type:        genai.TypeString,
								Description: "変更のスコープ（オプション）",
							},
							"description": {
								Type:        genai.TypeString,
								Description: "詳細な説明（オプション）",
							},
						},
						Required: []string{"message", "type"},
					},
				},
				{
					Name:        "gitDiff",
					Description: "Gitの差分を取得します",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {
								Type:        genai.TypeString,
								Description: "確認するパス（オプション）",
							},
							"staged": {
								Type:        genai.TypeBoolean,
								Description: "ステージングされた変更の差分を表示するかどうか",
							},
						},
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

	// メッセージの送信と応答の取得
	fmt.Printf("\nAIに送信するメッセージ:\n%s\n\n", userInput)
	resp, err := chat.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
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
						if p.Name == "execCommand" {
							args, ok := p.Args["args"].([]any)
							if !ok {
								args = []any{}
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
								Response: map[string]any{
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
						} else if p.Name == "getGitHubIssue" {
							// パラメータの取得
							repository := p.Args["repository"].(string)
							issueNumber := int(p.Args["issueNumber"].(float64))

							// ghコマンドを実行
							cmd := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", issueNumber), "--repo", repository, "--json", "title,body,state,labels,createdAt,updatedAt,assignees,milestone,comments")
							output, err := cmd.CombinedOutput()
							if err != nil {
								return fmt.Errorf("Issueの取得に失敗: %v\n出力: %s", err, output)
							}

							// 結果をFunctionResponseとして送信
							resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
								Name: "getGitHubIssue",
								Response: map[string]any{
									"raw": string(output),
								},
							})
							if err != nil {
								return fmt.Errorf("Issue情報の送信に失敗: %v", err)
							}

							// 続きのタスクを実行するために、ループを継続
							shouldBreak = false
						} else if p.Name == "gitStatus" {
							// パラメータの取得
							var path string
							if p.Args["path"] != nil {
								path = p.Args["path"].(string)
							}

							// git statusコマンドを実行
							cmd := exec.Command("git", "status")
							if path != "" {
								cmd.Args = append(cmd.Args, path)
							}
							output, err := cmd.CombinedOutput()
							if err != nil {
								return fmt.Errorf("Gitステータスの取得に失敗: %v\n出力: %s", err, output)
							}

							// 結果をFunctionResponseとして送信
							resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
								Name: "gitStatus",
								Response: map[string]any{
									"status": string(output),
								},
							})
							if err != nil {
								return fmt.Errorf("Gitステータスの送信に失敗: %v", err)
							}

							// 続きのタスクを実行するために、ループを継続
							shouldBreak = false
						} else if p.Name == "gitAdd" {
							// パラメータの取得
							var paths []any
							if p.Args["paths"] != nil {
								paths = p.Args["paths"].([]any)
							}
							all := p.Args["all"].(bool)

							// git addコマンドを実行
							cmd := exec.Command("git", "add")
							if all {
								cmd.Args = append(cmd.Args, "--all")
							} else if len(paths) > 0 {
								for _, path := range paths {
									cmd.Args = append(cmd.Args, path.(string))
								}
							} else {
								cmd.Args = append(cmd.Args, ".")
							}
							output, err := cmd.CombinedOutput()
							if err != nil {
								return fmt.Errorf("Git addに失敗: %v\n出力: %s", err, output)
							}

							// 結果をFunctionResponseとして送信
							resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
								Name: "gitAdd",
								Response: map[string]any{
									"success": err == nil,
									"output":  string(output),
									"error":   err,
								},
							})
							if err != nil {
								return fmt.Errorf("Git add情報の送信に失敗: %v", err)
							}

							// 続きのタスクを実行するために、ループを継続
							shouldBreak = false
						} else if p.Name == "gitCommit" {
							// パラメータの取得
							message := p.Args["message"].(string)
							commitType := p.Args["type"].(string)
							scope := p.Args["scope"].(string)
							description := p.Args["description"].(string)

							// git commitコマンドを実行
							cmd := exec.Command("git", "commit", "-m", message)
							if commitType != "" {
								cmd.Args = append(cmd.Args, "-m", fmt.Sprintf("%s: %s", commitType, message))
							}
							if scope != "" {
								cmd.Args = append(cmd.Args, "-m", fmt.Sprintf("(%s): %s", scope, description))
							}
							output, err := cmd.CombinedOutput()
							if err != nil {
								return fmt.Errorf("Git commitに失敗: %v\n出力: %s", err, output)
							}

							// 結果をFunctionResponseとして送信
							resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
								Name: "gitCommit",
								Response: map[string]any{
									"success": err == nil,
									"output":  string(output),
									"error":   err,
								},
							})
							if err != nil {
								return fmt.Errorf("Git commit情報の送信に失敗: %v", err)
							}

							// 続きのタスクを実行するために、ループを継続
							shouldBreak = false
						} else if p.Name == "gitDiff" {
							// パラメータの取得
							var path string
							if p.Args["path"] != nil {
								path = p.Args["path"].(string)
							}
							var staged bool
							if p.Args["staged"] != nil {
								staged = p.Args["staged"].(bool)
							}

							// git diffコマンドを実行
							cmd := exec.Command("git", "diff")
							if path != "" {
								cmd.Args = append(cmd.Args, path)
							}
							if staged {
								cmd.Args = append(cmd.Args, "--cached")
							}
							output, err := cmd.CombinedOutput()
							if err != nil {
								return fmt.Errorf("Git diffに失敗: %v\n出力: %s", err, output)
							}

							// 結果をFunctionResponseとして送信
							resp, err = chat.SendMessage(ctx, genai.FunctionResponse{
								Name: "gitDiff",
								Response: map[string]any{
									"diff": string(output),
								},
							})
							if err != nil {
								return fmt.Errorf("Git diff情報の送信に失敗: %v", err)
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
