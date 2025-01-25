package vim

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/cmd/roo-helper/models"
)

func TestNewBridge(t *testing.T) {
	bridge := NewBridge()
	if bridge == nil {
		t.Error("NewBridge() returned nil")
	}
}

func TestBridgeSendOutput(t *testing.T) {
	// 標準出力をキャプチャするための設定
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	bridge := NewBridge()
	testContent := "test content"
	err := bridge.SendOutput("test", testContent)

	// 標準出力を元に戻す
	w.Close()
	os.Stdout = oldStdout

	// 出力の検証
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("SendOutput() error = %v", err)
	}

	var result OutputFormat
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Failed to unmarshal output: %v", err)
	}

	if result.Type != "test" {
		t.Errorf("Expected type 'test', got %s", result.Type)
	}

	content, ok := result.Content.(string)
	if !ok {
		t.Error("Failed to assert content as string")
	}
	if content != testContent {
		t.Errorf("Expected content %s, got %s", testContent, content)
	}
}

func TestBridgeSendError(t *testing.T) {
	// 標準エラー出力をキャプチャするための設定
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	bridge := NewBridge()
	testError := errors.New("test error")
	err := bridge.SendError(testError)

	// 標準エラー出力を元に戻す
	w.Close()
	os.Stderr = oldStderr

	// 出力の検証
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("SendError() error = %v", err)
	}

	var result OutputFormat
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Failed to unmarshal output: %v", err)
	}

	if result.Type != "error" {
		t.Errorf("Expected type 'error', got %s", result.Type)
	}

	response, ok := result.Content.(map[string]interface{})
	if !ok {
		t.Error("Failed to assert content as Response")
	}
	if response["success"].(bool) != false {
		t.Error("Expected success to be false")
	}
	if response["error"].(string) != testError.Error() {
		t.Errorf("Expected error message %s, got %s", testError.Error(), response["error"])
	}
}

func TestParseVimInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "valid input",
			input: `{"command": "test", "input": "data"}`,
			want: map[string]interface{}{
				"command": "test",
				"input":   "data",
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   "invalid json",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVimInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVimInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVimInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVimInput(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"command": "test",
				"input":   "data",
			},
			wantErr: false,
		},
		{
			name: "missing command",
			input: map[string]interface{}{
				"input": "data",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVimInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVimInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridgeFormatResponse(t *testing.T) {
	// 標準出力をキャプチャするための設定
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	bridge := NewBridge()
	testResponse := models.Response{
		Success: true,
		Data:    "test data",
	}
	err := bridge.FormatResponse(testResponse)

	// 標準出力を元に戻す
	w.Close()
	os.Stdout = oldStdout

	// 出力の検証
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("FormatResponse() error = %v", err)
	}

	var result OutputFormat
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Failed to unmarshal output: %v", err)
	}

	if result.Type != "success" {
		t.Errorf("Expected type 'success', got %s", result.Type)
	}

	response, ok := result.Content.(map[string]interface{})
	if !ok {
		t.Error("Failed to assert content as Response")
	}
	if response["success"].(bool) != true {
		t.Error("Expected success to be true")
	}
	if response["data"].(string) != "test data" {
		t.Errorf("Expected data 'test data', got %s", response["data"])
	}
}