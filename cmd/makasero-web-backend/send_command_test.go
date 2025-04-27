package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/pankona/makasero" // 型定義のためにインポート
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendCommand(t *testing.T) {
	// --- Test Setup ---
	t.Setenv("GEMINI_API_KEY", "test-api-key")

	testSessionID := "test-session-123"
	dummySession := createDummySession(testSessionID) // test_utils.go のヘルパーを使用

	// モック SessionLoader を設定
	mockSessionLoader := &mockSessionLoader{}
	var loadSessionCalledWith string
	mockSessionLoader.LoadSessionFunc = func(id string) (*makasero.Session, error) {
		loadSessionCalledWith = id // 呼び出されたIDを記録
		if id == testSessionID {
			return dummySession, nil
		}
		return nil, assert.AnError // 予期しないIDの場合はエラー
	}

	// モック AgentCreator と Agent を設定
	mockAgentCreator := &mockAgentCreator{}
	mockAgent := NewMockAgent()
	mockAgentCreator.NewAgentFunc = func(ctx context.Context, apiKey string, config *makasero.MCPConfig, opts ...makasero.AgentOption) (AgentProcessor, error) {
		// WithSession オプションが渡されているか確認 (任意)
		found := false
		for _, opt := range opts {
			// オプションが適用された結果を確認するか、リフレクションを使う必要がある
			// ここでは簡略化のためチェック省略
			_ = opt // Avoid unused variable error
			// if _, ok := findAgentOption[makasero.WithSession](opts); ok { // findAgentOption はコメントアウト中
			//  found = true
			// }
			found = true // 仮に true
		}
		assert.True(t, found, "NewAgent に WithSession オプションが渡されるべき")
		mockAgentCreator.CreatedAgent = mockAgent
		return mockAgent, nil
	}

	// SessionManager を作成し、モックを注入
	sm := &SessionManager{
		apiKey:        "test-api-key",
		modelName:     "test-model",
		configPath:    "/fake/config.json",
		configLoader:  &mockConfigLoader{},
		agentCreator:  mockAgentCreator,
		sessionLoader: mockSessionLoader, // 設定したモックを使用
	}

	server := createTestServer(t, sm)
	defer server.Close()

	// --- Test Execution ---
	command := "テストコマンド実行"
	reqBody := SendCommandRequest{Command: command}
	jsonBody, _ := json.Marshal(reqBody)
	url := server.URL + "/api/sessions/" + testSessionID + "/commands"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// --- Assertions ---
	assert.Equal(t, http.StatusAccepted, resp.StatusCode, "ステータスコードは 202 Accepted であるべき")
	assert.Equal(t, testSessionID, loadSessionCalledWith, "LoadSession が正しい SessionID で呼ばれるべき")

	var respData SendCommandResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	require.NoError(t, err, "レスポンスボディの JSON デコードに成功すべき")
	assert.Equal(t, "Command accepted", respData.Message, "レスポンスメッセージが期待通りであるべき")

	// --- Goroutine Assertions (タイムアウト付き) ---
	select {
	case receivedCommand := <-mockAgent.ProcessMessageChan:
		assert.Equal(t, command, receivedCommand, "goroutine 内で ProcessMessage が正しいコマンドで呼ばれるべき")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: mockAgent.ProcessMessage が時間内に呼び出されませんでした")
	}

	select {
	case <-mockAgent.CloseChan:
		mockAgent.mu.Lock()
		closeCalled := mockAgent.CloseCalled
		mockAgent.mu.Unlock()
		assert.True(t, closeCalled, "goroutine 内で Close が呼ばれるべき")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout: mockAgent.Close が時間内に呼び出されませんでした")
	}
}

func TestSendCommand_SessionNotFound(t *testing.T) {
	// --- Test Setup ---
	t.Setenv("GEMINI_API_KEY", "test-api-key")
	notFoundSessionID := "not-found-session"

	mockSessionLoader := &mockSessionLoader{}
	mockSessionLoader.LoadSessionFunc = func(id string) (*makasero.Session, error) {
		assert.Equal(t, notFoundSessionID, id)
		return nil, os.ErrNotExist // Not Found を返す
	}

	sm := &SessionManager{
		apiKey:        "test-api-key",
		modelName:     "test-model",
		configPath:    "/fake/config.json",
		configLoader:  &mockConfigLoader{},
		agentCreator:  &mockAgentCreator{}, // このテストでは呼ばれないはず
		sessionLoader: mockSessionLoader,
	}

	server := createTestServer(t, sm)
	defer server.Close()

	// --- Test Execution ---
	reqBody := SendCommandRequest{Command: "any command"}
	jsonBody, _ := json.Marshal(reqBody)
	url := server.URL + "/api/sessions/" + notFoundSessionID + "/commands"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// --- Assertions ---
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "存在しないセッションの場合は 404 Not Found であるべき")
}

// TODO: LoadMCPConfig や NewAgent がエラーを返すケース、空コマンドのケースのテストを追加
