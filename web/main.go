package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Summary   string    `json:"summary"`
}

type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AddMessageRequest struct {
	Message string `json:"message"`
}

type CreateSessionRequest struct {
	Prompt string `json:"prompt"`
}

func main() {
	// 静的ファイルを提供
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/api/sessions", handleSessions)
	http.HandleFunc("/api/sessions/", handleSessionRequests)

	log.Printf("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleListSessions(w, r)
	case http.MethodPost:
		handleCreateSession(w, r)
	default:
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "METHOD_NOT_ALLOWED",
				Message: "Method not allowed",
			},
		}, http.StatusMethodNotAllowed)
	}
}

func handleListSessions(w http.ResponseWriter, r *http.Request) {
	// makasero -ls コマンドを実行
	cmd := exec.Command("makasero", "-ls")
	output, err := cmd.Output()
	if err != nil {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to execute makasero -ls: " + err.Error(),
			},
		}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, Response{
		Success: true,
		Data:    string(output),
	}, http.StatusOK)
}

func handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		}, http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Prompt is required",
			},
		}, http.StatusBadRequest)
		return
	}

	// makasero コマンドを実行して新しいセッションを作成
	cmd := exec.Command("makasero", req.Prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create session: " + string(output),
			},
		}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, Response{
		Success: true,
		Data:    string(output),
	}, http.StatusOK)
}

func handleSessionRequests(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if sessionID == "" {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Session ID is required",
			},
		}, http.StatusBadRequest)
		return
	}

	// メッセージ追加のエンドポイント
	if strings.HasSuffix(r.URL.Path, "/message") {
		sessionID = strings.TrimSuffix(sessionID, "/message")
		handleAddMessage(w, r, sessionID)
		return
	}

	// セッション詳細の取得
	if r.Method == http.MethodGet {
		handleSessionDetail(w, r, sessionID)
		return
	}

	writeJSON(w, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "Method not allowed",
		},
	}, http.StatusMethodNotAllowed)
}

func handleAddMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "METHOD_NOT_ALLOWED",
				Message: "Only POST method is allowed",
			},
		}, http.StatusMethodNotAllowed)
		return
	}

	var req AddMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		}, http.StatusBadRequest)
		return
	}

	// makasero -s コマンドを実行
	cmd := exec.Command("makasero", "-s", sessionID, req.Message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to execute makasero -s: " + string(output),
			},
		}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, Response{
		Success: true,
		Data:    string(output),
	}, http.StatusOK)
}

func handleSessionDetail(w http.ResponseWriter, r *http.Request, sessionID string) {
	// makasero -show コマンドを実行
	cmd := exec.Command("makasero", "-sh", sessionID)
	output, err := cmd.Output()
	if err != nil {
		writeJSON(w, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to execute makasero -show: " + err.Error(),
			},
		}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, Response{
		Success: true,
		Data:    string(output),
	}, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
