package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CreateSessionRequest struct {
	Prompt string `json:"prompt"`
}

type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

type SendCommandRequest struct {
	Command string `json:"command"`
}

type SendCommandResponse struct {
	Message string `json:"message"`
}

type SessionManager struct {
	makaseroCmd []string // 実行するコマンドと引数
}

func NewSessionManager() *SessionManager {
	cmdStr := os.Getenv("MAKASERO_COMMAND")
	var cmdArgs []string

	if cmdStr != "" {
		// 環境変数で指定されたコマンドをスペースで分割
		cmdArgs = strings.Fields(cmdStr)
	} else {
		// デフォルトのコマンド
		// 環境変数から取得できない場合のフォールバックを追加
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			// WindowsなどHOMEがない場合
			homeDir = os.Getenv("USERPROFILE")
		}
		// Note: ここで homeDir が空だと panic する可能性があるため、エラーハンドリングを追加してもよい
		makaseroRepoPath := filepath.Join(homeDir, "repos", "makasero") // 仮のパス、必要に応じて調整
		// cmd/makasero/main.go へのパスを組み立てる
		makaseroMainPath := filepath.Join(makaseroRepoPath, "cmd", "makasero", "main.go")
		cmdArgs = []string{"go", "run", makaseroMainPath}
	}

	log.Printf("Using command: %v", cmdArgs) // どのコマンドを使うかログ出力

	return &SessionManager{
		makaseroCmd: cmdArgs,
	}
}

func (sm *SessionManager) StartSession(prompt string) (string, error) {
	sessionID := uuid.New().String()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting user home directory: %v. Falling back to current directory for .makasero.", err)
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	makaseroDir := filepath.Join(homeDir, ".makasero")
	if err := os.MkdirAll(makaseroDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .makasero directory: %w", err)
	}

	sessionsDir := filepath.Join(makaseroDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create sessions directory: %w", err)
	}

	configPath := filepath.Join(makaseroDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := []byte(`{"mcpServers":{}}`)
		if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
			return "", fmt.Errorf("failed to create config file '%s': %w", configPath, err)
		}
	}

	log.Printf("Starting session with ID: %s", sessionID)
	log.Printf("Home directory: %s", homeDir)
	log.Printf("Config path: %s", configPath)
	log.Printf("Sessions directory: %s", sessionsDir)

	if len(sm.makaseroCmd) == 0 {
		return "", fmt.Errorf("makasero command is not configured in SessionManager")
	}
	baseArgs := make([]string, len(sm.makaseroCmd)-1)
	copy(baseArgs, sm.makaseroCmd[1:])
	args := append(baseArgs, "-debug", "-config", configPath, "-s", sessionID, prompt)
	cmd := exec.Command(sm.makaseroCmd[0], args...)

	if os.Getenv("MAKASERO_COMMAND") == "" {
		cmdDirHome := os.Getenv("HOME")
		if cmdDirHome == "" {
			cmdDirHome = os.Getenv("USERPROFILE")
		}
		if cmdDirHome != "" {
			repoDir := filepath.Join(cmdDirHome, "repos", "makasero")
			if _, err := os.Stat(repoDir); err == nil {
				cmd.Dir = repoDir
			} else {
				log.Printf("Warning: Default command directory '%s' not found, using current working directory.", repoDir)
				wd, _ := os.Getwd()
				cmd.Dir = wd
			}
		} else {
			wd, _ := os.Getwd()
			cmd.Dir = wd
			log.Printf("Warning: HOME or USERPROFILE not set, using current working directory '%s' for command execution", cmd.Dir)
		}
	} else {
		log.Printf("MAKASERO_COMMAND is set, not setting cmd.Dir explicitly.")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-2.0-flash-lite" // Default model
	}

	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env, "MODEL_NAME="+modelName)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start command '%s' with args %v in dir '%s': %v", sm.makaseroCmd[0], args, cmd.Dir, err)
		return "", fmt.Errorf("failed to start session process: %w", err)
	}

	log.Printf("Started process with PID: %d for session %s", cmd.Process.Pid, sessionID)

	if cmd.Process != nil {
		err = cmd.Process.Release()
		if err != nil {
			log.Printf("Warning: Failed to release process %d for session %s: %v", cmd.Process.Pid, sessionID, err)
		}
	} else {
		log.Printf("Warning: Command start for session %s succeeded but process is nil. Args: %v", sessionID, args)
	}

	return sessionID, nil
}

func (sm *SessionManager) SendCommand(sessionID, command string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting user home directory: %v. Falling back to current directory for .makasero.", err)
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	makaseroDir := filepath.Join(homeDir, ".makasero")
	if err := os.MkdirAll(makaseroDir, 0755); err != nil {
		return fmt.Errorf("failed to create .makasero directory: %w", err)
	}

	sessionsDir := filepath.Join(makaseroDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	configPath := filepath.Join(makaseroDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := []byte(`{"mcpServers":{}}`)
		if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
			return fmt.Errorf("failed to create config file '%s': %w", configPath, err)
		}
	}

	if len(sm.makaseroCmd) == 0 {
		return fmt.Errorf("makasero command is not configured in SessionManager")
	}
	baseArgs := make([]string, len(sm.makaseroCmd)-1)
	copy(baseArgs, sm.makaseroCmd[1:])
	args := append(baseArgs, "-debug", "-config", configPath, "-s", sessionID, command)
	cmd := exec.Command(sm.makaseroCmd[0], args...)

	if os.Getenv("MAKASERO_COMMAND") == "" {
		cmdDirHome := os.Getenv("HOME")
		if cmdDirHome == "" {
			cmdDirHome = os.Getenv("USERPROFILE")
		}
		if cmdDirHome != "" {
			repoDir := filepath.Join(cmdDirHome, "repos", "makasero")
			if _, err := os.Stat(repoDir); err == nil {
				cmd.Dir = repoDir
			} else {
				log.Printf("Warning: Default command directory '%s' not found, using current working directory.", repoDir)
				wd, _ := os.Getwd()
				cmd.Dir = wd
			}
		} else {
			wd, _ := os.Getwd()
			cmd.Dir = wd
			log.Printf("Warning: HOME or USERPROFILE not set, using current working directory '%s' for command execution", cmd.Dir)
		}
	} else {
		log.Printf("MAKASERO_COMMAND is set, not setting cmd.Dir explicitly.")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-2.0-flash-lite" // Default model
	}

	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env, "MODEL_NAME="+modelName)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start command '%s' with args %v in dir '%s': %v", sm.makaseroCmd[0], args, cmd.Dir, err)
		return fmt.Errorf("failed to send command process: %w", err)
	}

	log.Printf("Started process with PID: %d for sending command to session %s", cmd.Process.Pid, sessionID)

	if cmd.Process != nil {
		err = cmd.Process.Release()
		if err != nil {
			log.Printf("Warning: Failed to release process %d for session %s command: %v", cmd.Process.Pid, sessionID, err)
		}
	} else {
		log.Printf("Warning: Command start for session %s command succeeded but process is nil. Args: %v", sessionID, args)
	}

	return nil
}

func (sm *SessionManager) GetSessionStatus(sessionID string) ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	possiblePaths := []string{
		filepath.Join(homeDir, ".makasero", "sessions", sessionID+".json"),
		filepath.Join(os.Getenv("HOME"), "repos", "makasero", ".makasero", "sessions", sessionID+".json"),
	}

	maxRetries := 3
	retryDelay := 500 * time.Millisecond

	for _, sessionFilePath := range possiblePaths {
		log.Printf("Checking for session file at: %s", sessionFilePath)
		
		for i := 0; i < maxRetries; i++ {
			if _, err := os.Stat(sessionFilePath); err == nil {
				data, err := os.ReadFile(sessionFilePath)
				if err != nil {
					return nil, fmt.Errorf("failed to read session file '%s': %w", sessionFilePath, err)
				}
				return data, nil
			} else if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to check session file '%s': %w", sessionFilePath, err)
			}

			if i < maxRetries-1 {
				time.Sleep(retryDelay)
			}
		}
	}

	return nil, fmt.Errorf("session not found: %s", sessionID)
}

func handleCreateSession(w http.ResponseWriter, r *http.Request, sm *SessionManager) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	sessionID, err := sm.StartSession(req.Prompt)
	if err != nil {
		http.Error(w, "Failed to start session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateSessionResponse{
		SessionID: sessionID,
		Status: "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func handleSendCommand(w http.ResponseWriter, r *http.Request, sm *SessionManager, sessionID string) {
	var req SendCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	if err := sm.SendCommand(sessionID, req.Command); err != nil {
		http.Error(w, "Failed to send command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := SendCommandResponse{
		Message: "Command accepted",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetSessionStatus(w http.ResponseWriter, r *http.Request, sm *SessionManager, sessionID string) {
	data, err := sm.GetSessionStatus(sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "session not found") {
			http.Error(w, fmt.Sprintf("Session not found: %s", sessionID), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get session status: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	sessionManager := NewSessionManager()

	http.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCreateSession(w, r, sessionManager)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		pathSegments := strings.Split(r.URL.Path, "/")
		if len(pathSegments) < 4 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		sessionID := pathSegments[3]

		if len(pathSegments) == 4 {
			if r.Method == http.MethodGet {
				handleGetSessionStatus(w, r, sessionManager, sessionID)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(pathSegments) == 5 && pathSegments[4] == "commands" {
			if r.Method == http.MethodPost {
				handleSendCommand(w, r, sessionManager, sessionID)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Invalid path", http.StatusBadRequest)
		}
	})

	handler := corsMiddleware(http.DefaultServeMux)

	log.Printf("Starting server on :%s", *port)
	if err := http.ListenAndServe(":"+*port, handler); err != nil {
		log.Fatal(err)
	}
}
