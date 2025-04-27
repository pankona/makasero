package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/pankona/makasero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSession(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-api-key")

	mockAgentCreator := &mockAgentCreator{}
	mockAgent := NewMockAgent()
	mockAgent.SessionID = "test-session-id"
	mockAgent.ProcessedPrompt = "test prompt"
	mockAgentCreator.NewAgentFunc = func(ctx context.Context, apiKey string, config *makasero.MCPConfig, opts ...makasero.AgentOption) (AgentProcessor, error) {
		mockAgentCreator.CreatedAgent = mockAgent
		return mockAgent, nil
	}

	sm := &SessionManager{
		apiKey:        "test-api-key",
		modelName:     "test-model",
		configPath:    "/fake/config.json",
		configLoader:  &mockConfigLoader{},
		agentCreator:  mockAgentCreator,
		sessionLoader: &mockSessionLoader{},
	}

	server := createTestServer(t, sm)
	defer server.Close()

	prompt := "テストプロンプト"
	reqBody := CreateSessionRequest{Prompt: prompt}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", server.URL+"/api/sessions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode, "ステータスコードは 202 Accepted であるべき")

	var respData CreateSessionResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	require.NoError(t, err, "レスポンスボディの JSON デコードに成功すべき")

	assert.NotEmpty(t, respData.SessionID, "レスポンスには session_id が含まれるべき")
	assert.Equal(t, "accepted", respData.Status, "レスポンスの status は accepted であるべき")

	select {
	case receivedPrompt := <-mockAgent.ProcessMessageChan:
		assert.Equal(t, prompt, receivedPrompt, "goroutine 内で ProcessMessage が正しいプロンプトで呼ばれるべき")
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

	require.NotNil(t, mockAgentCreator.CreatedAgent, "AgentCreator が Agent を生成しているべき")
	require.Equal(t, mockAgent.SessionID, "test-session-id", "SessionIDが正しく設定されているべき")
	require.Equal(t, mockAgent.ProcessedPrompt, "test prompt", "ProcessedPromptが正しく設定されているべき")
}

func TestCreateSession_EmptyPrompt(t *testing.T) {
	sm := &SessionManager{
		apiKey:        "test-api-key",
		modelName:     "test-model",
		configPath:    "/fake/config.json",
		configLoader:  &mockConfigLoader{},
		agentCreator:  &mockAgentCreator{},
		sessionLoader: &mockSessionLoader{},
	}
	server := createTestServer(t, sm)
	defer server.Close()

	reqBody := CreateSessionRequest{Prompt: ""}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", server.URL+"/api/sessions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "空プロンプトの場合は 400 Bad Request であるべき")
}
