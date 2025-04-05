package tools

import (
	"fmt"
	"testing"
)

func TestExecCommand_Name(t *testing.T) {
	cmd := &ExecCommand{}
	if got := cmd.Name(); got != "execCommand" {
		t.Errorf("ExecCommand.Name() = %v, want %v", got, "execCommand")
	}
}

func TestExecCommand_Description(t *testing.T) {
	cmd := &ExecCommand{}
	if got := cmd.Description(); got == "" {
		t.Error("ExecCommand.Description() returned empty string")
	}
}

func TestExecCommand_Execute(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid command",
			args: map[string]interface{}{
				"command": "echo 'test'",
			},
			wantErr: false,
		},
		{
			name: "invalid command",
			args: map[string]interface{}{
				"command": "nonexistentcommand",
			},
			wantErr: true,
		},
		{
			name:    "missing command",
			args:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "wrong type",
			args: map[string]interface{}{
				"command": 123,
			},
			wantErr: true,
		},
	}

	cmd := &ExecCommand{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cmd.Execute(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("ExecCommand.Execute() returned empty output for valid command")
			}
		})
	}
}

func Example_execCommand() {
	cmd := &ExecCommand{}
	output, err := cmd.Execute(map[string]interface{}{
		"command": "echo 'Hello, World!'",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Print(output)
	// Output: Hello, World!
}
