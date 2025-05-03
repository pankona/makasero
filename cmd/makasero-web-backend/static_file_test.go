package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticFileServing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "static-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	htmlContent := []byte("<html><body><h1>Test Page</h1></body></html>")
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), htmlContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cssDir := filepath.Join(tempDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatalf("Failed to create css directory: %v", err)
	}
	cssContent := []byte("body { color: red; }")
	if err := os.WriteFile(filepath.Join(cssDir, "style.css"), cssContent, 0644); err != nil {
		t.Fatalf("Failed to write css file: %v", err)
	}

	mux := http.NewServeMux()
	
	fs := http.FileServer(http.Dir(tempDir))
	mux.Handle("/", fs)
	
	server := httptest.NewServer(mux)
	defer server.Close()

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   []byte
		expectedType   string
	}{
		{
			name:           "Root path should serve index.html",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   htmlContent,
			expectedType:   "text/html",
		},
		{
			name:           "Direct file access should work",
			path:           "/index.html",
			expectedStatus: http.StatusOK,
			expectedBody:   htmlContent,
			expectedType:   "text/html",
		},
		{
			name:           "Subdirectory file access should work",
			path:           "/css/style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   cssContent,
			expectedType:   "text/css",
		},
		{
			name:           "Non-existent file should return 404",
			path:           "/not-found.html",
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
			expectedType:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + tc.path)
			if err != nil {
				t.Fatalf("Failed to get %s: %v", tc.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedBody != nil {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				if string(body) != string(tc.expectedBody) {
					t.Errorf("Expected body %s, got %s", string(tc.expectedBody), string(body))
				}
			}

			if tc.expectedType != "" && !strings.Contains(resp.Header.Get("Content-Type"), tc.expectedType) {
				t.Errorf("Expected Content-Type to contain %s, got %s", tc.expectedType, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestAPIAndStaticIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "static-api-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	htmlContent := []byte("<html><body><h1>Test Page</h1></body></html>")
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), htmlContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mainMux := http.NewServeMux()
	mainMux.Handle("/api/", apiMux)
	
	fs := http.FileServer(http.Dir(tempDir))
	mainMux.Handle("/", fs)
	
	handler := corsMiddleware(mainMux)
	
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("API endpoint should work", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/test")
		if err != nil {
			t.Fatalf("Failed to get API endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expected := `{"status":"ok"}`
		if string(body) != expected {
			t.Errorf("Expected body %s, got %s", expected, string(body))
		}

		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected CORS header to be set for API endpoint")
		}
	})

	t.Run("Static file should work", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/index.html")
		if err != nil {
			t.Fatalf("Failed to get static file: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if string(body) != string(htmlContent) {
			t.Errorf("Expected body %s, got %s", string(htmlContent), string(body))
		}

		if resp.Header.Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("Expected no CORS header for static file, got %s", resp.Header.Get("Access-Control-Allow-Origin"))
		}
	})
}
