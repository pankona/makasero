package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ChatSession はチャットセッションの情報を保持します
type ChatSession struct {
	ID        string        `json:"id"`
	StartTime time.Time     `json:"start_time"`
	FilePath  string        `json:"file_path,omitempty"`
	Messages  []ChatMessage `json:"messages"`
}

// SaveSession はチャットセッションを保存します
func (s *ChatSession) SaveSession(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("セッションディレクトリの作成に失敗: %w", err)
	}

	filename := filepath.Join(dir, fmt.Sprintf("session_%s.json", s.ID))
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("セッションのJSONエンコードに失敗: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("セッションの保存に失敗: %w", err)
	}

	return nil
}

// LoadSession は指定されたIDのチャットセッションを読み込みます
func LoadSession(dir, id string) (*ChatSession, error) {
	filename := filepath.Join(dir, fmt.Sprintf("session_%s.json", id))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("セッションの読み込みに失敗: %w", err)
	}

	var session ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("セッションのJSONデコードに失敗: %w", err)
	}

	return &session, nil
}

// ListSessions はセッションの一覧を取得します
func ListSessions(dir string) ([]ChatSession, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("セッションディレクトリの読み込みに失敗: %w", err)
	}

	var sessions []ChatSession
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				continue
			}

			var session ChatSession
			if err := json.Unmarshal(data, &session); err != nil {
				continue
			}

			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}
