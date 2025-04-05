package agent

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/pankona/makasero/tools"
)

func TestIntegration_ProcessMessage(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY is not set")
	}

	tests := []struct {
		name            string
		input           string
		wantErr         bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "empty_input",
			input:           "",
			wantErr:         true,
			wantContains:    []string{},
			wantNotContains: []string{},
		},
		{
			name:  "pwd_command",
			input: "現在のディレクトリを表示して",
			wantContains: []string{
				"現在", "ディレクトリ", "パス",
				"/home/pankona/go/src/github.com/pankona/makasero",
			},
			wantNotContains: []string{
				"Error", "error",
			},
		},
		{
			name:  "show_current_directory",
			input: "現在のディレクトリを表示して",
			wantContains: []string{
				"現在のディレクトリは",
				"/home/pankona/go/src/github.com/pankona/makasero",
				"です",
			},
			wantNotContains: []string{
				"Error", "error", "エラー",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeminiClient(apiKey)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}
			defer client.Close()

			// ツールを登録
			client.RegisterTool(&tools.ExecCommand{})
			client.RegisterTool(&tools.CompleteTool{})

			got, err := client.ProcessMessage(context.Background(), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ProcessMessage() error = %v", err)
				return
			}

			// 期待する文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("ProcessMessage() = %v, want containing %v", got, want)
				}
			}

			// 期待しない文字列が含まれていないか確認
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("ProcessMessage() = %v, does not want containing %v", got, notWant)
				}
			}

			t.Logf("Response: %s", got)
		})
	}
}
