package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func setupTestSessionManager(t *testing.T) *SessionManager {
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
	return &SessionManager{
		makaseroCmd: append([]string{echoCmdPath}, echoArgs...),
	}
}

func SetupFakeHomeDir(t *testing.T, tempDir string) {
	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir)
}

func SetupTestEnvironment(t *testing.T, tempDir string) (string, string, string) {
	SetupFakeHomeDir(t, tempDir)
	
	makaseroDir := filepath.Join(tempDir, ".makasero")
	sessionsDir := filepath.Join(makaseroDir, "sessions")
	configPath := filepath.Join(makaseroDir, "config.json")
	
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
	}
	
	defaultConfig := []byte(`{"mcpServers":{}}`)
	if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
		t.Fatalf("テスト用設定ファイルの作成に失敗: %v", err)
	}
	
	return makaseroDir, sessionsDir, configPath
}
