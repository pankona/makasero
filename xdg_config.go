package makasero

import (
	"os"
	"path/filepath"
)

// GetConfigDir returns the configuration directory following XDG Base Directory specification.
// It first checks for existing legacy config at ~/.makasero, then falls back to XDG_CONFIG_HOME or ~/.config.
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Check for existing legacy configuration directory
	legacyDir := filepath.Join(homeDir, ".makasero")
	if _, err := os.Stat(legacyDir); err == nil {
		// Legacy directory exists, continue using it for backward compatibility
		return legacyDir, nil
	}

	// Use XDG Base Directory specification
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(xdgConfigHome, "makasero"), nil
}

// GetSessionsDir returns the sessions directory based on the config directory.
func GetSessionsDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sessions"), nil
}

// GetConfigFilePath returns the path to the main configuration file.
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}