package tools

import (
	"fmt"
	"os/exec"
)

// ExecCommand はシェルコマンドを実行するツールです
type ExecCommand struct{}

// Name はツールの名前を返します
func (e *ExecCommand) Name() string {
	return "execCommand"
}

// Description はツールの説明を返します
func (e *ExecCommand) Description() string {
	return "Execute a shell command and return its output"
}

// Execute はシェルコマンドを実行し、結果を返します
func (e *ExecCommand) Execute(args map[string]interface{}) (string, error) {
	cmd, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command argument is required")
	}

	output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v", err)
	}

	return string(output), nil
}
