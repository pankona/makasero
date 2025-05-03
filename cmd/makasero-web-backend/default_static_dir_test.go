package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultStaticDirectory(t *testing.T) {
	tempHomeDir, err := os.MkdirTemp("", "home-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}
	defer os.RemoveAll(tempHomeDir)

	defaultStaticDir := filepath.Join(tempHomeDir, ".makasero", "web-frontend")
	if err := os.MkdirAll(defaultStaticDir, 0755); err != nil {
		t.Fatalf("Failed to create default static directory: %v", err)
	}
	
	indexContent := []byte("<html><body><h1>Makasero Web Frontend</h1><p>This directory is ready to serve static frontend files.</p></body></html>")
	indexPath := filepath.Join(defaultStaticDir, "index.html")
	if err := os.WriteFile(indexPath, indexContent, 0644); err != nil {
		t.Fatalf("Failed to create index.html: %v", err)
	}
	
	fs := http.FileServer(http.Dir(defaultStaticDir))
	server := httptest.NewServer(fs)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to get index page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedContent := string(indexContent)
	if string(body) != expectedContent {
		t.Errorf("Expected body %s, got %s", expectedContent, string(body))
	}
}
