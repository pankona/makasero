package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupTestTaskManager(t *testing.T) *TaskManager {
	t.Setenv("GEMINI_API_KEY", "test-api-key")

	var echoCmdPath string
	var echoArgs []string
	var err error

	if runtime.GOOS == "windows" {
		echoCmdPath, err = exec.LookPath("cmd")
		if err != nil {
			t.Fatalf("Could not find 'cmd' command: %v", err)
		}
		echoArgs = []string{"/c", "echo", "dummy"}
	} else {
		echoCmdPath, err = exec.LookPath("echo")
		if err != nil {
			t.Fatalf("Could not find 'echo' command: %v", err)
		}
		echoArgs = []string{"dummy"}
	}
	return &TaskManager{
		makaseroCmd: append([]string{echoCmdPath}, echoArgs...),
	}
}

func TestHandleCreateTask(t *testing.T) {
	tempDir := t.TempDir()
	makaseroDir := filepath.Join(tempDir, ".makasero")
	sessionsDir := filepath.Join(makaseroDir, "sessions")
	configPath := filepath.Join(makaseroDir, "config.json")

	setupFakeHomeDir := func(t *testing.T) {
		t.Setenv("HOME", tempDir)
		t.Setenv("USERPROFILE", tempDir)
	}

	tests := []struct {
		name           string
		requestBody    CreateTaskRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder, string)
	}{
		{
			name: "正常なリクエスト (echo)",
			requestBody: CreateTaskRequest{
				Prompt: "テストプロンプト",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, expectedTaskID string) {
				var response CreateTaskResponse
				bodyBytes := w.Body.Bytes()
				if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&response); err != nil {
					t.Errorf("レスポンスのデコードに失敗: %v. Body: %s", err, string(bodyBytes))
					return
				}
				if response.TaskID == "" {
					t.Error("TaskIDが空です")
				}
				if response.Status != "pending" {
					t.Errorf("期待されるステータス 'pending' に対して、実際は '%s' でした", response.Status)
				}
			},
		},
		{
			name: "空のプロンプト",
			requestBody: CreateTaskRequest{
				Prompt: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, expectedTaskID string) {
				expectedMsg := "Prompt is required\n"
				if w.Body.String() != expectedMsg {
					t.Errorf("期待されるエラーメッセージ '%s' と異なります: '%s'", expectedMsg, w.Body.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupFakeHomeDir(t)
			if err := os.MkdirAll(sessionsDir, 0755); err != nil {
				t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
			}
			defaultConfig := []byte(`{"mcpServers":{}}`)
			if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
				t.Fatalf("テスト用設定ファイルの作成に失敗: %v", err)
			}

			tm := setupTestTaskManager(t)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()
			handleCreateTask(rr, req, tm)

			var generatedTaskID string
			if rr.Code == http.StatusCreated {
				var resp CreateTaskResponse
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&resp); err == nil {
					generatedTaskID = resp.TaskID
				} else {
					t.Logf("正常リクエストのはずがレスポンスのデコードに失敗: %v", err)
				}
			}

			if rr.Code != tt.expectedStatus {
				t.Errorf("期待されるステータスコード %d に対して、実際は %d でした. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr, generatedTaskID)
			}
		})
	}
}

func TestHandleSendCommand(t *testing.T) {
	tempDir := t.TempDir()
	makaseroDir := filepath.Join(tempDir, ".makasero")
	sessionsDir := filepath.Join(makaseroDir, "sessions")
	configPath := filepath.Join(makaseroDir, "config.json")

	setupFakeHomeDir := func(t *testing.T) {
		t.Setenv("HOME", tempDir)
		t.Setenv("USERPROFILE", tempDir)
	}

	tests := []struct {
		name           string
		taskID         string
		requestBody    SendCommandRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "正常なコマンド (echo)",
			taskID: "test-task-id-send",
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
			name:   "空のコマンド",
			taskID: "test-task-id-send-empty",
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
			setupFakeHomeDir(t)
			if err := os.MkdirAll(sessionsDir, 0755); err != nil {
				t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
			}
			defaultConfig := []byte(`{"mcpServers":{}}`)
			if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
				t.Fatalf("テスト用設定ファイルの作成に失敗: %v", err)
			}

			tm := setupTestTaskManager(t)

			body, _ := json.Marshal(tt.requestBody)
			url := "/api/tasks/" + tt.taskID + "/commands"
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handleSendCommand(w, req, tm, tt.taskID)

			if w.Code != tt.expectedStatus {
				t.Errorf("期待されるステータスコード %d に対して、実際は %d でした. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestHandleGetTaskStatus(t *testing.T) {
	tempDir := t.TempDir()
	makaseroDir := filepath.Join(tempDir, ".makasero")
	sessionsDir := filepath.Join(makaseroDir, "sessions")
	configPath := filepath.Join(makaseroDir, "config.json")

	setupFakeHomeDir := func(t *testing.T) {
		t.Setenv("HOME", tempDir)
		t.Setenv("USERPROFILE", tempDir)
	}

	tests := []struct {
		name             string
		taskID           string
		setupSessionFile func(t *testing.T)
		expectedStatus   int
		checkResponse    func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "存在しないタスク",
			taskID: "non-existent-task",
			setupSessionFile: func(t *testing.T) {
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				expectedMsgPrefix := "Task not found:"
				bodyStr := w.Body.String()
				if !strings.HasPrefix(strings.TrimSpace(bodyStr), expectedMsgPrefix) {
					t.Errorf("期待されるエラーメッセージの接頭辞 '%s' と異なります: '%s'", expectedMsgPrefix, bodyStr)
				}
			},
		},
		{
			name:   "存在するタスク",
			taskID: "existing-task",
			setupSessionFile: func(t *testing.T) {
				sessionFilePath := filepath.Join(sessionsDir, "existing-task.json")
				dummyData := `{"id": "existing-task", "status": "running"}`
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
				if id, ok := data["id"].(string); !ok || id != "existing-task" {
					t.Errorf("期待されるID 'existing-task' がレスポンスに含まれていません: %v", data)
				}
				if status, ok := data["status"].(string); !ok || status != "running" {
					t.Errorf("期待されるステータス 'running' がレスポンスに含まれていません: %v", data)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupFakeHomeDir(t)
			if err := os.MkdirAll(sessionsDir, 0755); err != nil {
				t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
			}
			defaultConfig := []byte(`{"mcpServers":{}}`)
			if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
				t.Fatalf("テスト用設定ファイルの作成に失敗: %v", err)
			}
			tt.setupSessionFile(t)

			tm := setupTestTaskManager(t)

			req := httptest.NewRequest(http.MethodGet, "/api/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()

			handleGetTaskStatus(w, req, tm, tt.taskID)

			if w.Code != tt.expectedStatus {
				t.Errorf("期待されるステータスコード %d に対して、実際は %d でした. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
