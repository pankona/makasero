package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pankona/makasero"
)

// --- makasero パッケージのモック ---

// モック用の AgentProcessor 実装 (mockAgent は AgentProcessor を満たす)
type mockAgent struct {
	CloseFunc          func() error
	ProcessMessageFunc func(ctx context.Context, userInput string) error
	// GetSessionFunc     func() *makasero.Session // 不要なら削除
	mu                 sync.Mutex
	ProcessMessageArgs []string
	CloseCalled        bool
	ProcessCalled      bool        // ProcessMessage が呼ばれたか
	ProcessMessageChan chan string // ProcessMessage 呼び出し通知用 (goroutineテスト用)
	CloseChan          chan bool   // Close 呼び出し通知用 (goroutineテスト用)
	SessionID          string      // テスト用のセッションID
	ProcessedPrompt    string      // テスト用の処理されたプロンプト
}

// NewMockAgent はテスト用のモックエージェントを作成
func NewMockAgent() *mockAgent {
	return &mockAgent{
		ProcessMessageChan: make(chan string, 1), // バッファ付きチャネル
		CloseChan:          make(chan bool, 1),
	}
}

func (m *mockAgent) Close() error {
	m.mu.Lock()
	m.CloseCalled = true
	m.mu.Unlock()
	if m.CloseFunc != nil {
		err := m.CloseFunc()
		m.CloseChan <- true // 呼び出しを通知
		return err
	}
	m.CloseChan <- true // 呼び出しを通知
	return nil
}

func (m *mockAgent) ProcessMessage(ctx context.Context, userInput string) error {
	m.mu.Lock()
	m.ProcessCalled = true
	m.ProcessMessageArgs = append(m.ProcessMessageArgs, userInput)
	m.mu.Unlock()

	var err error
	if m.ProcessMessageFunc != nil {
		err = m.ProcessMessageFunc(ctx, userInput)
	}
	// 結果に関わらずチャネルに通知 (goroutine テスト用)
	m.ProcessMessageChan <- userInput

	return err
}

// --- モック用のインターフェース実装 ---

type mockConfigLoader struct {
	LoadMCPConfigFunc func(path string) (*makasero.MCPConfig, error)
}

func (m *mockConfigLoader) LoadMCPConfig(path string) (*makasero.MCPConfig, error) {
	if m.LoadMCPConfigFunc != nil {
		return m.LoadMCPConfigFunc(path)
	}
	// デフォルトのダミー実装
	return &makasero.MCPConfig{MCPServers: map[string]makasero.MCPServerConfig{}}, nil
}

type mockAgentCreator struct {
	NewAgentFunc func(ctx context.Context, apiKey string, config *makasero.MCPConfig, opts ...makasero.AgentOption) (AgentProcessor, error)
	// テストで生成された Agent を保持・取得できるようにする
	CreatedAgent AgentProcessor
}

func (m *mockAgentCreator) NewAgent(ctx context.Context, apiKey string, config *makasero.MCPConfig, opts ...makasero.AgentOption) (AgentProcessor, error) {
	if m.NewAgentFunc != nil {
		agent, err := m.NewAgentFunc(ctx, apiKey, config, opts...)
		m.CreatedAgent = agent // 生成された Agent を保持
		return agent, err
	}
	// デフォルトでは mockAgent を返す
	mock := NewMockAgent()
	m.CreatedAgent = mock
	return mock, nil
}

type mockSessionLoader struct {
	LoadSessionFunc func(id string) (*makasero.Session, error)
}

func (m *mockSessionLoader) LoadSession(id string) (*makasero.Session, error) {
	if m.LoadSessionFunc != nil {
		return m.LoadSessionFunc(id)
	}
	// デフォルトでは Not Found
	return nil, os.ErrNotExist
}

// --- テストユーティリティ関数 ---

// setupTestSessionManager はモック実装を受け取れるように変更 (または不要になる可能性)
// テストごとに SessionManager を直接作成し、フィールドにモックをセットする方が柔軟かもしれない
func setupTestSessionManager(t *testing.T, configPath string, cl ConfigLoader, ac AgentCreator, sl SessionLoader) *SessionManager {
	t.Setenv("GEMINI_API_KEY", "test-api-key")
	apiKey := os.Getenv("GEMINI_API_KEY")
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-1.5-flash-latest"
	}
	if configPath == "" {
		tempDir := t.TempDir()
		_, _, configPath = SetupTestEnvironment(t, tempDir)
	}

	// デフォルトモックを設定
	if cl == nil {
		cl = &mockConfigLoader{}
	}
	if ac == nil {
		ac = &mockAgentCreator{}
	}
	if sl == nil {
		sl = &mockSessionLoader{}
	}

	return &SessionManager{
		apiKey:        apiKey,
		modelName:     modelName,
		configPath:    configPath,
		configLoader:  cl,
		agentCreator:  ac,
		sessionLoader: sl,
	}
}

// SetupFakeHomeDir は変更なし
func SetupFakeHomeDir(t *testing.T, tempDir string) {
	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir)
}

// SetupTestEnvironment は変更なし
func SetupTestEnvironment(t *testing.T, tempDir string) (string, string, string) {
	SetupFakeHomeDir(t, tempDir)

	makaseroDir := filepath.Join(tempDir, ".makasero")
	sessionsDir := filepath.Join(makaseroDir, "sessions")
	configPath := filepath.Join(makaseroDir, "config.json")

	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
	}

	defaultConfig := []byte(`{"mcpServers":{}}`)
	if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
		t.Fatalf("テスト用設定ファイルの作成に失敗: %v", err)
	}

	return makaseroDir, sessionsDir, configPath
}

// --- テストでよく使うヘルパー ---

// createTestServer は SessionManager を受け取る
func createTestServer(t *testing.T, sm *SessionManager) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCreateSession(w, r, sm)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathSegments) < 3 {
			http.Error(w, "Invalid API path", http.StatusBadRequest)
			return
		}
		sessionID := pathSegments[2]

		if len(pathSegments) == 3 && r.Method == http.MethodGet {
			handleGetSessionStatus(w, r, sm, sessionID)
		} else if len(pathSegments) == 4 && pathSegments[3] == "commands" && r.Method == http.MethodPost {
			handleSendCommand(w, r, sm, sessionID)
		} else {
			if len(pathSegments) == 3 {
				http.Error(w, "Method not allowed for /api/sessions/{sessionID}", http.StatusMethodNotAllowed)
			} else if len(pathSegments) == 4 && pathSegments[3] == "commands" {
				http.Error(w, "Method not allowed for /api/sessions/{sessionID}/commands", http.StatusMethodNotAllowed)
			} else {
				http.Error(w, fmt.Sprintf("Invalid path under /api/sessions/%s", sessionID), http.StatusBadRequest)
			}
		}
	})
	return httptest.NewServer(mux)
}

// findAgentOption はコメントアウト (型アサーションが機能しないため)
/*
func findAgentOption[T any](opts []makasero.AgentOption) (T, bool) {
	for _, opt := range opts {
		// makasero.AgentOption が func(*Agent) のような型だと仮定すると、
		// この型アサーションは機能しない
		// if v, ok := opt.(T); ok {
		// 	return v, true
		// }

		// 代わりにリフレクションを使うか、オプションの適用結果を確認する必要がある
	}
	var zero T
	return zero, false
}
*/

// --- ダミーのセッションデータ作成 ---
func createDummySession(id string) *makasero.Session {
	return &makasero.Session{
		ID:        id,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
		SerializedHistory: []*makasero.SerializableContent{
			{
				Role: "user",
				Parts: []makasero.SerializablePart{
					{Type: "text", Content: "最初のプロンプト"},
				},
			},
			{
				Role: "model",
				Parts: []makasero.SerializablePart{
					{Type: "text", Content: "応答"},
				},
			},
		},
	}
}

// セッションファイルを書き込むヘルパー
func writeSessionFile(t *testing.T, sessionsDir string, session *makasero.Session) {
	filePath := filepath.Join(sessionsDir, session.ID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal dummy session: %v", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write dummy session file %s: %v", filePath, err)
	}
}
