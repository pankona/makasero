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

func TestHandleCreateSession(t *testing.T) {
	tempDir := t.TempDir()
	makaseroDir := filepath.Join(tempDir, ".makasero")
	sessionsDir := filepath.Join(makaseroDir, "sessions")
	configPath := filepath.Join(makaseroDir, "config.json")

	tests := []struct {
		name           string
		requestBody    CreateSessionRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder, string)
	}{
		{
			name: "正常なリクエスト (echo)",
			requestBody: CreateSessionRequest{
				Prompt: "テストプロンプト",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, expectedSessionID string) {
				var response CreateSessionResponse
				bodyBytes := w.Body.Bytes()
				if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&response); err != nil {
					t.Errorf("レスポンスのデコードに失敗: %v. Body: %s", err, string(bodyBytes))
					return
				}
				if response.SessionID == "" {
					t.Error("SessionIDが空です")
				}
				if response.Status != "pending" {
					t.Errorf("期待されるステータス 'pending' に対して、実際は '%s' でした", response.Status)
				}
			},
		},
		{
			name: "空のプロンプト",
			requestBody: CreateSessionRequest{
				Prompt: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, expectedSessionID string) {
				expectedMsg := "Prompt is required\n"
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
			req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()
			handleCreateSession(rr, req, sm)

			var generatedSessionID string
			if rr.Code == http.StatusCreated {
				var resp CreateSessionResponse
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&resp); err == nil {
					generatedSessionID = resp.SessionID
				} else {
					t.Logf("正常リクエストのはずがレスポンスのデコードに失敗: %v", err)
				}
			}

			if rr.Code != tt.expectedStatus {
				t.Errorf("期待されるステータスコード %d に対して、実際は %d でした. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr, generatedSessionID)
			}
		})
	}
}
