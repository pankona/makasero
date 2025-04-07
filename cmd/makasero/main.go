package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"os/exec"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	sessionDir = ".makasero/sessions"
)

type Session struct {
	ID                string                 `json:"id"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	History           []*genai.Content       `json:"-"` // JSON化しない
	SerializedHistory []*SerializableContent `json:"history"`
}

type SerializableContent struct {
	Parts []SerializablePart `json:"parts"`
	Role  string             `json:"role"`
}

type SerializablePart struct {
	Type    string `json:"type"`    // "text", "function_call", "function_response" など
	Content any    `json:"content"` // 実際のデータ
}

func (s *Session) MarshalJSON() ([]byte, error) {
	// History を SerializedHistory に変換
	s.SerializedHistory = make([]*SerializableContent, len(s.History))
	for i, content := range s.History {
		serialized := &SerializableContent{
			Role:  content.Role,
			Parts: make([]SerializablePart, len(content.Parts)),
		}

		for j, part := range content.Parts {
			switch p := part.(type) {
			case genai.Text:
				serialized.Parts[j] = SerializablePart{
					Type:    "text",
					Content: string(p),
				}
			case genai.FunctionCall:
				serialized.Parts[j] = SerializablePart{
					Type:    "function_call",
					Content: p,
				}
			case genai.FunctionResponse:
				serialized.Parts[j] = SerializablePart{
					Type:    "function_response",
					Content: p,
				}
			}
		}
		s.SerializedHistory[i] = serialized
	}

	// 一時的な構造体を作成してマーシャル
	type Alias Session
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(s)})
}

func (s *Session) UnmarshalJSON(data []byte) error {
	// 一時的な構造体を作成してアンマーシャル
	type Alias Session
	aux := &struct{ *Alias }{Alias: (*Alias)(s)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// SerializedHistory を History に変換
	s.History = make([]*genai.Content, len(s.SerializedHistory))
	for i, serialized := range s.SerializedHistory {
		content := &genai.Content{
			Role:  serialized.Role,
			Parts: make([]genai.Part, len(serialized.Parts)),
		}

		for j, part := range serialized.Parts {
			switch part.Type {
			case "text":
				content.Parts[j] = genai.Text(part.Content.(string))
			case "function_call":
				fc := part.Content.(map[string]interface{})
				content.Parts[j] = genai.FunctionCall{
					Name: fc["Name"].(string),
					Args: fc["Args"].(map[string]interface{}),
				}
			case "function_response":
				fr := part.Content.(map[string]interface{})
				content.Parts[j] = genai.FunctionResponse{
					Name:     fr["Name"].(string),
					Response: fr["Response"].(map[string]interface{}),
				}
			}
		}
		s.History[i] = content
	}

	return nil
}

func loadSession(id string) (*Session, error) {
	path := filepath.Join(sessionDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func saveSession(session *Session) error {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(sessionDir, session.ID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func listSessions() error {
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("セッションはありません")
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			id := strings.TrimSuffix(entry.Name(), ".json")
			session, err := loadSession(id)
			if err != nil {
				fmt.Printf("セッション %s の読み込みに失敗: %v\n", id, err)
				continue
			}
			fmt.Printf("Session ID: %s\n", session.ID)
			fmt.Printf("Created: %s\n", session.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Messages: %d\n", len(session.History))

			// 初期プロンプト（ユーザーからの最初のメッセージ）を表示
			if len(session.History) > 0 {
				for _, content := range session.History {
					if content.Role == "user" {
						fmt.Printf("初期プロンプト: ")
						for _, part := range content.Parts {
							if text, ok := part.(genai.Text); ok {
								// 長すぎる場合は省略
								prompt := string(text)
								if len(prompt) > 100 {
									prompt = prompt[:97] + "..."
								}
								fmt.Printf("%s\n", prompt)
								break
							}
						}
						break
					}
				}
			}

			fmt.Println()
		}
	}
	return nil
}

func generateSessionID() string {
	timestamp := time.Now().Format("20060102150405")
	random := make([]byte, 4)
	rand.Read(random)
	return fmt.Sprintf("%s_%x", timestamp, random)
}

var (
	debug            = flag.Bool("debug", false, "デバッグモード")
	promptFile       = flag.String("f", "", "プロンプトファイル")
	listSessionsFlag = flag.Bool("ls", false, "利用可能なセッション一覧を表示")
	sessionID        = flag.String("s", "", "継続するセッションID")
	showHistory      = flag.String("sh", "", "指定したセッションIDの会話履歴全文を表示")
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

func showSessionHistory(id string) error {
	session, err := loadSession(id)
	if err != nil {
		return fmt.Errorf("セッション %s の読み込みに失敗: %v", id, err)
	}

	fmt.Printf("セッションID: %s\n", session.ID)
	fmt.Printf("作成日時: %s\n", session.CreatedAt.Format(time.RFC3339))
	fmt.Printf("最終更新: %s\n", session.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("メッセージ数: %d\n\n", len(session.History))

	for i, content := range session.History {
		fmt.Printf("--- メッセージ %d ---\n", i+1)
		fmt.Printf("役割: %s\n", content.Role)
		for _, part := range content.Parts {
			switch p := part.(type) {
			case genai.Text:
				fmt.Printf("%s\n", string(p))
			case genai.FunctionCall:
				fmt.Printf("関数呼び出し: %s\n", p.Name)
				fmt.Printf("引数: %+v\n", p.Args)
			case genai.FunctionResponse:
				fmt.Printf("関数レスポンス: %s\n", p.Name)
				fmt.Printf("結果: %+v\n", p.Response)
			}
		}
		fmt.Println()
	}
	return nil
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
							message := ""
							if p.Args["message"] != nil {
								message = p.Args["message"].(string)
							}
							commitType := ""
							if p.Args["type"] != nil {
								commitType = p.Args["type"].(string)
							}
							scope := ""
							if p.Args["scope"] != nil {
								scope = p.Args["scope"].(string)
							}
							description := ""
							if p.Args["description"] != nil {
								description = p.Args["description"].(string)
							}

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
							// 会話履歴を更新
							session.History = chat.History
							session.UpdatedAt = time.Now()
							return saveSession(session)
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
