package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pankona/makasero" // 型定義のためにインポート
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetSessionStatus(t *testing.T) {
	tempDir := t.TempDir()
	_, sessionsDir, _ := SetupTestEnvironment(t, tempDir)

	tests := []struct {
		name             string
		sessionID        string
		setupSessionFile func(t *testing.T, sessionsDir string)
		expectedStatus   int
		checkResponse    func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "存在しないセッション",
			sessionID: "non-existent-session",
			setupSessionFile: func(t *testing.T, sessionsDir string) {
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				expectedMsgPrefix := "Session not found:"
				bodyStr := w.Body.String()
				if !strings.HasPrefix(strings.TrimSpace(bodyStr), expectedMsgPrefix) {
					t.Errorf("期待されるエラーメッセージの接頭辞 '%s' と異なります: '%s'", expectedMsgPrefix, bodyStr)
				}
			},
		},
		{
			name:      "存在するセッション",
			sessionID: "existing-session",
			setupSessionFile: func(t *testing.T, sessionsDir string) {
				sessionFilePath := filepath.Join(sessionsDir, "existing-session.json")
				dummyData := `{"id": "existing-session", "status": "running"}`
				if err := os.WriteFile(sessionFilePath, []byte(dummyData), 0644); err != nil {
					t.Fatalf("テスト用セッションファイルの作成に失敗: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var data map[string]interface{}
				bodyBytes := w.Body.Bytes()
				if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&data); err != nil {
					t.Errorf("レスポンスのJSONデコードに失敗: %v. Body: %s", err, string(bodyBytes))
					return
				}
				if id, ok := data["id"].(string); !ok || id != "existing-session" {
					t.Errorf("期待されるID 'existing-session' がレスポンスに含まれていません: %v", data)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupSessionFile(t, sessionsDir)

			// カスタムのmockSessionLoaderを作成
			sessionLoader := &mockSessionLoader{}
			sessionLoader.LoadSessionFunc = func(id string) (*makasero.Session, error) {
				if id == "existing-session" {
					// JSONテキストを直接パースして、JSONとして返すようにする
					jsonData := []byte(`{"id": "existing-session", "status": "running"}`)
					var session makasero.Session
					if err := json.Unmarshal(jsonData, &session); err != nil {
						t.Fatalf("Session JSONのパースに失敗: %v", err)
					}
					return &session, nil
				}
				return nil, os.ErrNotExist
			}

			sm := setupTestSessionManager(t, "", nil, nil, sessionLoader)

			req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+tt.sessionID, nil)
			w := httptest.NewRecorder()

			handleGetSessionStatus(w, req, sm, tt.sessionID)

			if w.Code != tt.expectedStatus {
				t.Errorf("期待されるステータスコード %d に対して、実際は %d でした. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestGetSessionStatus_Found(t *testing.T) {
	// --- Test Setup ---
	testSessionID := "get-session-test-found"
	dummySession := createDummySession(testSessionID)

	// モック SessionLoader を設定
	mockSessionLoader := &mockSessionLoader{}
	var loadSessionCalledWith string
	mockSessionLoader.LoadSessionFunc = func(id string) (*makasero.Session, error) {
		loadSessionCalledWith = id
		if id == testSessionID {
			return dummySession, nil
		}
		return nil, os.ErrNotExist // 予期しない ID は Not Found
	}

	// SessionManager を作成し、モックを注入
	sm := &SessionManager{
		apiKey:        "dummy", // このテストでは使われない
		modelName:     "dummy",
		configPath:    "dummy",
		configLoader:  &mockConfigLoader{},
		agentCreator:  &mockAgentCreator{},
		sessionLoader: mockSessionLoader, // 設定したモックを使用
	}

	server := createTestServer(t, sm)
	defer server.Close()

	// --- Test Execution ---
	url := server.URL + "/api/sessions/" + testSessionID
	req, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// --- Assertions ---
	assert.Equal(t, http.StatusOK, resp.StatusCode, "ステータスコードは 200 OK であるべき")
	assert.Equal(t, testSessionID, loadSessionCalledWith, "LoadSession が正しい SessionID で呼ばれるべき")

	var respData makasero.Session
	err = json.NewDecoder(resp.Body).Decode(&respData)
	require.NoError(t, err, "レスポンスボディの JSON デコードに成功すべき")

	// 返されたセッションデータの内容を検証
	assert.Equal(t, dummySession.ID, respData.ID)
	// Time の比較は .Equal() だとずれる可能性があるので注意 (必要なら time.Time.Equal() や誤差許容)
	assert.WithinDuration(t, dummySession.CreatedAt, respData.CreatedAt, time.Second)
	assert.WithinDuration(t, dummySession.UpdatedAt, respData.UpdatedAt, time.Second)
	// History は JSON に含まれないはず (SerializedHistory を確認)
	// assert.Nil(t, respData.History, "デコード後の History は nil のはず") // 実際には空のスライスになる場合もある
	// SerializedHistory の内容を比較 (構造が深いので注意)
	require.Equal(t, len(dummySession.SerializedHistory), len(respData.SerializedHistory), "SerializedHistory の長さが一致すべき")
	if len(dummySession.SerializedHistory) == len(respData.SerializedHistory) { // nil pointer dereference 防止
		for i := range dummySession.SerializedHistory {
			assert.Equal(t, dummySession.SerializedHistory[i].Role, respData.SerializedHistory[i].Role)
			require.Equal(t, len(dummySession.SerializedHistory[i].Parts), len(respData.SerializedHistory[i].Parts))
			if len(dummySession.SerializedHistory[i].Parts) == len(respData.SerializedHistory[i].Parts) {
				// Parts の内容も比較 (型アサーションが必要になる場合がある)
				// ここでは簡易的に Type のみを比較
				for j := range dummySession.SerializedHistory[i].Parts {
					assert.Equal(t, dummySession.SerializedHistory[i].Parts[j].Type, respData.SerializedHistory[i].Parts[j].Type)
					// Content の比較は型によって方法が変わる
				}
			}
		}
	}
}

func TestGetSessionStatus_NotFound(t *testing.T) {
	// --- Test Setup ---
	testSessionID := "get-session-test-not-found"

	// モック SessionLoader を設定 (Not Found を返す)
	mockSessionLoader := &mockSessionLoader{}
	var loadSessionCalledWith string
	mockSessionLoader.LoadSessionFunc = func(id string) (*makasero.Session, error) {
		loadSessionCalledWith = id
		assert.Equal(t, testSessionID, id)
		return nil, os.ErrNotExist
	}

	// SessionManager を作成し、モックを注入
	sm := &SessionManager{
		apiKey:        "dummy",
		modelName:     "dummy",
		configPath:    "dummy",
		configLoader:  &mockConfigLoader{},
		agentCreator:  &mockAgentCreator{},
		sessionLoader: mockSessionLoader,
	}

	server := createTestServer(t, sm)
	defer server.Close()

	// --- Test Execution ---
	url := server.URL + "/api/sessions/" + testSessionID
	req, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// --- Assertions ---
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "ステータスコードは 404 Not Found であるべき")
	assert.Equal(t, testSessionID, loadSessionCalledWith, "LoadSession が正しい SessionID で呼ばれるべき")
}

// TODO: LoadSession が os.ErrNotExist 以外のエラーを返すケースのテストを追加
