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
				if status, ok := data["status"].(string); !ok || status != "running" {
					t.Errorf("期待されるステータス 'running' がレスポンスに含まれていません: %v", data)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupSessionFile(t, sessionsDir)

			sm := setupTestSessionManager(t)

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
