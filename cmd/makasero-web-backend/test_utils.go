package main

import (
	"os/exec"
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
