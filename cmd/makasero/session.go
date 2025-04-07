package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
)

const (
	sessionDir = ".makasero/sessions"
)

var (
	listSessionsFlag = flag.Bool("ls", false, "利用可能なセッション一覧を表示")
	sessionID        = flag.String("s", "", "継続するセッションID")
	showHistory      = flag.String("sh", "", "指定したセッションIDの会話履歴全文を表示")
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
