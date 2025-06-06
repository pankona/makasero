package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pankona/makasero"
	"github.com/pankona/makasero/mlog"
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
	apiKey        string
	modelName     string
	configPath    string
	configLoader  ConfigLoader
	agentCreator  AgentCreator
	sessionLoader SessionLoader
}

func NewSessionManager() (*SessionManager, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-1.5-flash-latest"
		log.Printf("MODEL_NAME not set, using default: %s", modelName)
	}

	_, configPath, _, err := setupMakaseroEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to setup makasero environment: %w", err)
	}
	log.Printf("Using config path: %s", configPath)

	return &SessionManager{
		apiKey:        apiKey,
		modelName:     modelName,
		configPath:    configPath,
		configLoader:  &defaultConfigLoader{},
		agentCreator:  &defaultAgentCreator{},
		sessionLoader: &defaultSessionLoader{},
	}, nil
}

func setupMakaseroEnvironment() (homeDir, configPath, sessionsDir string, err error) {
	homeDir, err = os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting user home directory: %v. Falling back to current directory for .makasero.", err)
		return "", "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Use XDG Base Directory specification with backward compatibility
	makaseroDir, err := makasero.GetConfigDir()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get config directory: %w", err)
	}

	if err := os.MkdirAll(makaseroDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create makasero directory: %w", err)
	}

	sessionsDir = filepath.Join(makaseroDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create sessions directory: %w", err)
	}

	configPath = filepath.Join(makaseroDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file '%s' not found, creating default.", configPath)
		defaultConfig := []byte(`{"mcpServers":{"claude":{"command":"claude","args":["mcp","serve"],"env":{}}}}`)
		if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
			return "", "", "", fmt.Errorf("failed to create config file '%s': %w", configPath, err)
		}
	} else if err != nil {
		return "", "", "", fmt.Errorf("failed to check config file '%s': %w", configPath, err)
	}

	return homeDir, configPath, sessionsDir, nil
}

func handleListSessions(w http.ResponseWriter, r *http.Request, sm *SessionManager) {
	sessions, err := makasero.ListSessions()
	if err != nil {
		log.Printf("Failed to list sessions: %v", err)
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}

	// セッションが見つからない場合でも空のリストを返す (エラーではない)
	if sessions == nil {
		sessions = []*makasero.Session{} // 空のスライスを初期化
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		log.Printf("Error encoding session list: %v", err)
		// ここまで来てエンコードに失敗したら Internal Server Error
		http.Error(w, "Failed to encode session list", http.StatusInternalServerError)
	}
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

	sessionID := uuid.New().String()
	ctx := context.Background()

	config, err := sm.configLoader.LoadMCPConfig(sm.configPath)
	if err != nil {
		log.Printf("Error loading MCP config from %s: %v", sm.configPath, err)
		http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
		return
	}

	opts := []makasero.AgentOption{
		makasero.WithCustomSessionID(sessionID),
		makasero.WithModelName(sm.modelName),
	}

	agentProcessor, err := sm.agentCreator.NewAgent(ctx, sm.apiKey, config, opts...)
	if err != nil {
		log.Printf("Failed to create agent for session %s: %v", sessionID, err)
		http.Error(w, "Failed to initialize session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		gCtx := context.Background()
		gLogger := log.New(os.Stderr, "[makasero-session-"+sessionID+"] ", log.LstdFlags|log.Lshortfile)
		gLogger.Printf("Starting background processing for session %s", sessionID)

		if err := agentProcessor.ProcessMessage(gCtx, req.Prompt); err != nil {
			mlog.Errorf(gCtx, "Error processing message for session %s: %v", sessionID, err)
		} else {
			mlog.Infof(gCtx, "Successfully finished processing for session %s", sessionID)
		}
		if err := agentProcessor.Close(); err != nil {
			mlog.Errorf(gCtx, "Error closing agent for session %s: %v", sessionID, err)
		}
	}()

	resp := CreateSessionResponse{
		SessionID: sessionID,
		Status:    "accepted",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error writing response for session %s: %v", sessionID, err)
	}
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

	ctx := context.Background()

	loadedSession, err := sm.sessionLoader.LoadSession(sessionID)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Session not found: %s", sessionID), http.StatusNotFound)
		} else {
			log.Printf("Failed to load session %s: %v", sessionID, err)
			http.Error(w, "Failed to load session data", http.StatusInternalServerError)
		}
		return
	}

	config, err := sm.configLoader.LoadMCPConfig(sm.configPath)
	if err != nil {
		log.Printf("Error loading MCP config from %s: %v", sm.configPath, err)
		http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
		return
	}

	opts := []makasero.AgentOption{
		makasero.WithSession(loadedSession),
		makasero.WithModelName(sm.modelName),
	}

	agentProcessor, err := sm.agentCreator.NewAgent(ctx, sm.apiKey, config, opts...)
	if err != nil {
		log.Printf("Failed to create agent for session %s command: %v", sessionID, err)
		http.Error(w, "Failed to initialize session for command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		gCtx := context.Background()
		gLogger := log.New(os.Stderr, "[makasero-cmd-"+sessionID+"] ", log.LstdFlags|log.Lshortfile)
		gLogger.Printf("Starting background command processing for session %s", sessionID)

		if err := agentProcessor.ProcessMessage(gCtx, req.Command); err != nil {
			mlog.Errorf(gCtx, "Error processing command for session %s: %v", sessionID, err)
		} else {
			mlog.Infof(gCtx, "Successfully finished command processing for session %s", sessionID)
		}
		if err := agentProcessor.Close(); err != nil {
			mlog.Errorf(gCtx, "Error closing agent for session %s command: %v", sessionID, err)
		}
	}()

	resp := SendCommandResponse{
		Message: "Command accepted",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error writing command response for session %s: %v", sessionID, err)
	}
}

func handleGetSessionStatus(w http.ResponseWriter, r *http.Request, sm *SessionManager, sessionID string) {
	sessionData, err := sm.sessionLoader.LoadSession(sessionID)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Session not found: %s", sessionID), http.StatusNotFound)
		} else {
			log.Printf("Failed to load session %s: %v", sessionID, err)
			http.Error(w, "Failed to get session status: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal session data for %s: %v", sessionID, err)
		http.Error(w, "Failed to serialize session data", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setupDefaultStaticDir attempts to set up and use the default static directory
func setupDefaultStaticDir() (string, error) {
	// Use XDG Base Directory specification with backward compatibility
	makaseroDir, err := makasero.GetConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	
	defaultStaticDir := filepath.Join(makaseroDir, "web-frontend")
	
	if _, err := os.Stat(defaultStaticDir); err == nil {
		return defaultStaticDir, nil
	}
	
	if err := os.MkdirAll(defaultStaticDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create default static directory: %w", err)
	}
	
	indexContent := []byte("<html><body><h1>Makasero Web Frontend</h1><p>This directory is ready to serve static frontend files.</p></body></html>")
	indexPath := filepath.Join(defaultStaticDir, "index.html")
	if err := os.WriteFile(indexPath, indexContent, 0644); err != nil {
		return defaultStaticDir, fmt.Errorf("created directory but failed to create placeholder index.html: %w", err)
	}
	
	return defaultStaticDir, nil
}

func main() {
	port := flag.String("port", "3000", "Port to listen on")
	staticDir := flag.String("static-dir", "", "Directory containing static files to serve")
	flag.Parse()

	log.SetPrefix("[makasero-backend] ")

	sessionManager, err := NewSessionManager()
	if err != nil {
		log.Fatalf("Failed to initialize SessionManager: %v", err)
	}

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListSessions(w, r, sessionManager)
		} else if r.Method == http.MethodPost {
			handleCreateSession(w, r, sessionManager)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	apiMux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		// 完全一致 "/api/sessions/" かどうかをチェック
		if r.URL.Path == "/api/sessions/" {
			if r.Method == http.MethodGet {
				handleListSessions(w, r, sessionManager)
			} else {
				http.Error(w, "Method not allowed for /api/sessions/", http.StatusMethodNotAllowed)
			}
			return // 完全一致の場合は以降の処理をスキップ
		}

		// "/api/sessions/{sessionID}" や "/api/sessions/{sessionID}/commands" の処理
		pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(pathSegments) < 3 {
			// このケースは上の完全一致チェックで処理されるはずだが、念のため
			http.Error(w, "Invalid API path structure near /api/sessions/", http.StatusBadRequest)
			return
		}

		if pathSegments[0] != "api" || pathSegments[1] != "sessions" {
			http.Error(w, "Invalid API path structure", http.StatusBadRequest)
			return
		}

		sessionID := pathSegments[2]

		if len(pathSegments) == 3 {
			if r.Method == http.MethodGet {
				handleGetSessionStatus(w, r, sessionManager, sessionID)
			} else {
				http.Error(w, "Method not allowed for /api/sessions/{sessionID}", http.StatusMethodNotAllowed)
			}
		} else if len(pathSegments) == 4 && pathSegments[3] == "commands" {
			if r.Method == http.MethodPost {
				handleSendCommand(w, r, sessionManager, sessionID)
			} else {
				http.Error(w, "Method not allowed for /api/sessions/{sessionID}/commands", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, fmt.Sprintf("Invalid path under /api/sessions/%s", sessionID), http.StatusBadRequest)
		}
	})

	mainMux := http.NewServeMux()

	mainMux.Handle("/api/", apiMux)

	if *staticDir == "" {
		defaultDir, err := setupDefaultStaticDir()
		if err != nil {
			log.Fatalf("Failed to set up default static directory: %v", err)
		}
		*staticDir = defaultDir
	}
	
	if _, err := os.Stat(*staticDir); os.IsNotExist(err) {
		log.Fatalf("Static directory '%s' does not exist", *staticDir)
	}
	
	log.Printf("Serving static files from: %s", *staticDir)
	fs := http.FileServer(http.Dir(*staticDir))
	
	mainMux.Handle("/", fs)

	handler := corsMiddleware(mainMux)

	log.Printf("Starting server on :%s", *port)
	if err := http.ListenAndServe(":"+*port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
