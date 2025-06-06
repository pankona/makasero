package makasero

import (
	"os"
	"path/filepath"
)

// GetConfigDir returns the configuration directory following XDG Base Directory specification.
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
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