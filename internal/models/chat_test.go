package models

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChatSession_SaveSession(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		session *ChatSession
		wantErr bool
	}{
		{
			name: "正常系：セッションの保存",
			session: &ChatSession{
				ID:        "test-session",
				StartTime: time.Now(),
				FilePath:  "test.go",
				Messages: []ChatMessage{
					{Role: "user", Content: "こんにちは"},
					{Role: "assistant", Content: "はい、こんにちは"},
				},
			},
			wantErr: false,
		},
		{
			name: "正常系：空のメッセージリスト",
			session: &ChatSession{
				ID:        "empty-session",
				StartTime: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.SaveSession(tmpDir)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// ファイルが作成されたことを確認
				filename := filepath.Join(tmpDir, fmt.Sprintf("session_%s.json", tt.session.ID))
				_, err := os.Stat(filename)
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadSession(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// テスト用のセッションを作成
	testSession := &ChatSession{
		ID:        "test-session",
		StartTime: time.Now(),
		FilePath:  "test.go",
		Messages: []ChatMessage{
			{Role: "user", Content: "こんにちは"},
			{Role: "assistant", Content: "はい、こんにちは"},
		},
	}
	err := testSession.SaveSession(tmpDir)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		dir     string
		id      string
		want    *ChatSession
		wantErr bool
	}{
		{
			name:    "正常系：セッションの読み込み",
			dir:     tmpDir,
			id:      "test-session",
			want:    testSession,
			wantErr: false,
		},
		{
			name:    "異常系：存在しないセッション",
			dir:     tmpDir,
			id:      "nonexistent",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "異常系：不正なディレクトリ",
			dir:     "/nonexistent",
			id:      "test-session",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadSession(tt.dir, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.FilePath, got.FilePath)
				assert.Equal(t, len(tt.want.Messages), len(got.Messages))
			}
		})
	}
}

func TestListSessions(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// テスト用のセッションを作成
	testSessions := []*ChatSession{
		{
			ID:        "session1",
			StartTime: time.Now(),
			FilePath:  "test1.go",
			Messages: []ChatMessage{
				{Role: "user", Content: "こんにちは1"},
			},
		},
		{
			ID:        "session2",
			StartTime: time.Now(),
			FilePath:  "test2.go",
			Messages: []ChatMessage{
				{Role: "user", Content: "こんにちは2"},
			},
		},
	}

	for _, session := range testSessions {
		err := session.SaveSession(tmpDir)
		assert.NoError(t, err)
	}

	tests := []struct {
		name    string
		dir     string
		want    int
		wantErr bool
	}{
		{
			name:    "正常系：セッション一覧の取得",
			dir:     tmpDir,
			want:    2,
			wantErr: false,
		},
		{
			name:    "正常系：空のディレクトリ",
			dir:     filepath.Join(tmpDir, "empty"),
			want:    0,
			wantErr: false,
		},
		{
			name:    "異常系：不正なディレクトリ",
			dir:     "/nonexistent",
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListSessions(tt.dir)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, len(got))
			}
		})
	}
}
