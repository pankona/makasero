package makasero

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	// Store original environment variables
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDGConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name              string
		xdgConfigHome     string
		expectedPath      string
	}{
		{
			name:            "with XDG_CONFIG_HOME set",
			xdgConfigHome:   "/tmp/custom-config",
			expectedPath:    "/tmp/custom-config/makasero",
		},
		{
			name:            "without XDG_CONFIG_HOME",
			xdgConfigHome:   "",
			expectedPath:    filepath.Join(homeDir, ".config", "makasero"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up XDG_CONFIG_HOME environment
			if tt.xdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			configDir, err := GetConfigDir()
			if err != nil {
				t.Fatalf("GetConfigDir() failed: %v", err)
			}

			if configDir != tt.expectedPath {
				t.Errorf("GetConfigDir() = %q, want %q", configDir, tt.expectedPath)
			}
		})
	}
}

func TestGetSessionsDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	// Clear XDG_CONFIG_HOME
	os.Unsetenv("XDG_CONFIG_HOME")

	sessionsDir, err := GetSessionsDir()
	if err != nil {
		t.Fatalf("GetSessionsDir() failed: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".config", "makasero", "sessions")
	if sessionsDir != expectedPath {
		t.Errorf("GetSessionsDir() = %q, want %q", sessionsDir, expectedPath)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	// Clear XDG_CONFIG_HOME
	os.Unsetenv("XDG_CONFIG_HOME")

	configFilePath, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath() failed: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".config", "makasero", "config.json")
	if configFilePath != expectedPath {
		t.Errorf("GetConfigFilePath() = %q, want %q", configFilePath, expectedPath)
	}
}