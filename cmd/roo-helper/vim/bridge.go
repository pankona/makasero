package vim

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rooveterinaryinc/hello-vim-plugin-2/cmd/roo-helper/models"
)

// OutputFormat defines the format of messages sent to Vim
type OutputFormat struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// Bridge handles communication between Go and Vim
type Bridge struct {
	stdout *json.Encoder
	stderr *json.Encoder
}

// NewBridge creates a new Bridge instance
func NewBridge() *Bridge {
	return &Bridge{
		stdout: json.NewEncoder(os.Stdout),
		stderr: json.NewEncoder(os.Stderr),
	}
}

// SendOutput sends formatted output to Vim
func (b *Bridge) SendOutput(outputType string, content interface{}) error {
	output := OutputFormat{
		Type:    outputType,
		Content: content,
	}
	return b.stdout.Encode(output)
}

// SendError sends an error message to Vim
func (b *Bridge) SendError(err error) error {
	output := OutputFormat{
		Type: "error",
		Content: models.Response{
			Success: false,
			Error:   err.Error(),
		},
	}
	return b.stderr.Encode(output)
}

// FormatResponse formats an API response for Vim
func (b *Bridge) FormatResponse(resp models.Response) error {
	outputType := "success"
	if !resp.Success {
		outputType = "error"
	}
	return b.SendOutput(outputType, resp)
}

// ParseVimInput parses input received from Vim
func ParseVimInput(input string) (map[string]interface{}, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input received")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, fmt.Errorf("failed to parse Vim input: %w", err)
	}

	return data, nil
}

// ValidateVimInput validates the required fields in Vim input
func ValidateVimInput(input map[string]interface{}) error {
	required := []string{"command"}
	for _, field := range required {
		if _, ok := input[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}