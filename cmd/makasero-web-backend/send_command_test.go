package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleSendCommand(t *testing.T) {
	tempDir := t.TempDir()
	makaseroDir := filepath.Join(tempDir, ".makasero")
	sessionsDir := filepath.Join(makaseroDir, "sessions")
	configPath := filepath.Join(makaseroDir, "config.json")

	tests := []struct {
		name           string
		sessionID      string
		requestBody    SendCommandRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "正常なコマンド (echo)",
			sessionID: "test-session-id-send",
			requestBody: SendCommandRequest{
				Command: "テストコマンド",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response SendCommandResponse
				bodyBytes := w.Body.Bytes()
				if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&response); err != nil {
					t.Errorf("レスポンスのデコードに失敗: %v. Body: %s", err, string(bodyBytes))
					return
				}
				expectedMsg := "Command accepted"
				if response.Message != expectedMsg {
					t.Errorf("期待されるメッセージ '%s' に対して、実際は '%s' でした", expectedMsg, response.Message)
				}
			},
		},
		{
			name:      "空のコマンド",
			sessionID: "test-session-id-send-empty",
			requestBody: SendCommandRequest{
				Command: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				expectedMsg := "Command is required\n"
				if w.Body.String() != expectedMsg {
					t.Errorf("期待されるエラーメッセージ '%s' と異なります: '%s'", expectedMsg, w.Body.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupFakeHomeDir(t, tempDir)
			if err := os.MkdirAll(sessionsDir, 0755); err != nil {
				t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
			}
			defaultConfig := []byte(`{"mcpServers":{}}`)
			if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
				t.Fatalf("テスト用設定ファイルの作成に失敗: %v", err)
			}

			sm := setupTestSessionManager(t)

			body, _ := json.Marshal(tt.requestBody)
			url := "/api/sessions/" + tt.sessionID + "/commands"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handleSendCommand(w, req, sm, tt.sessionID)

			if w.Code != tt.expectedStatus {
				t.Errorf("期待されるステータスコード %d に対して、実際は %d でした. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
