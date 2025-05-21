package makasero

import (
	"context"
	"fmt"
	"os" // Added import for os package
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// mockExecCommand is a mock for exec.Command
var mockExecCommand func(command string, args ...string) *exec.Cmd

// realExecCommand stores the original exec.Command
var realExecCommand = exec.Command

// helper function to mock exec.Command
func setupMockExecCommand(t *testing.T, expectedCommand string, expectedArgs []string, output string, err error) {
	mockExecCommand = func(command string, args ...string) *exec.Cmd {
		if command != expectedCommand {
			t.Errorf("Expected command '%s', got '%s'", expectedCommand, command)
		}
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("Expected args %v, got %v", expectedArgs, args)
		}
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(realExecCommand("go").Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"STDOUT=" + output,
		}
		if err != nil {
			cmd.Env = append(cmd.Env, "STDERR="+err.Error())
		}
		return cmd
	}
	execCommand = mockExecCommand
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for TestGhIssueCreate*.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	fmt.Fprint(os.Stderr, os.Getenv("STDERR"))

	// Simulate exit code based on whether STDERR is set
	if os.Getenv("STDERR") != "" {
		os.Exit(1)
	}
}

func teardownMockExecCommand() {
	execCommand = realExecCommand
}

// This is needed to allow execCommand to be replaced in tests
var execCommand = realExecCommand
// We need to modify the main functions.go file to use this execCommand variable.
// This will be done in a separate step.

func TestHandleGhIssueCreate_SuccessTitleOnly(t *testing.T) {
	expectedOutput := "https://github.com/owner/repo/issues/123"
	setupMockExecCommand(t, "gh", []string{"issue", "create", "--title", "Test Title"}, expectedOutput, nil)
	defer teardownMockExecCommand()

	args := map[string]any{
		"title": "Test Title",
	}
	result, err := handleGhIssueCreate(context.Background(), args)

	if err != nil {
		t.Fatalf("handleGhIssueCreate returned an unexpected error: %v", err)
	}
	if result["is_error"].(bool) {
		t.Errorf("Expected is_error to be false, got true. Output: %s", result["output"])
	}
	if result["output"] != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, result["output"])
	}
}

func TestHandleGhIssueCreate_SuccessTitleAndBody(t *testing.T) {
	expectedOutput := "https://github.com/owner/repo/issues/124"
	setupMockExecCommand(t, "gh", []string{"issue", "create", "--title", "Test Title", "--body", "Test Body"}, expectedOutput, nil)
	defer teardownMockExecCommand()

	args := map[string]any{
		"title": "Test Title",
		"body":  "Test Body",
	}
	result, err := handleGhIssueCreate(context.Background(), args)

	if err != nil {
		t.Fatalf("handleGhIssueCreate returned an unexpected error: %v", err)
	}
	if result["is_error"].(bool) {
		t.Errorf("Expected is_error to be false, got true. Output: %s", result["output"])
	}
	if result["output"] != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, result["output"])
	}
}

func TestHandleGhIssueCreate_SuccessTitleBodyAndRepo(t *testing.T) {
	expectedOutput := "https://github.com/testowner/testrepo/issues/1"
	setupMockExecCommand(t, "gh", []string{"issue", "create", "--title", "Test Title", "--body", "Test Body", "--repo", "testowner/testrepo"}, expectedOutput, nil)
	defer teardownMockExecCommand()

	args := map[string]any{
		"title": "Test Title",
		"body":  "Test Body",
		"repo":  "testowner/testrepo",
	}
	result, err := handleGhIssueCreate(context.Background(), args)

	if err != nil {
		t.Fatalf("handleGhIssueCreate returned an unexpected error: %v", err)
	}
	if result["is_error"].(bool) {
		t.Errorf("Expected is_error to be false, got true. Output: %s", result["output"])
	}
	if result["output"] != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, result["output"])
	}
}

func TestHandleGhIssueCreate_ErrorMissingTitle(t *testing.T) {
	args := map[string]any{
		"body": "Test Body",
	}
	result, err := handleGhIssueCreate(context.Background(), args)

	if err != nil {
		t.Fatalf("handleGhIssueCreate returned an unexpected error: %v", err)
	}
	if !result["is_error"].(bool) {
		t.Errorf("Expected is_error to be true, got false")
	}
	expectedErrorMsg := "title is required and cannot be empty"
	if result["output"] != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, result["output"])
	}
}

func TestHandleGhIssueCreate_ErrorGhCommandFails(t *testing.T) {
	simulatedError := fmt.Errorf("gh command failed")
	simulatedOutput := "Error: Some gh CLI error"
	// The combined output will be in the "output" field of the result when cmd.CombinedOutput() is used
	expectedFullErrorMsg := fmt.Sprintf("gh issue create failed: %v\nOutput: %s", simulatedError, simulatedOutput)

	setupMockExecCommand(t, "gh", []string{"issue", "create", "--title", "Test Title"}, simulatedOutput, simulatedError)
	defer teardownMockExecCommand()

	args := map[string]any{
		"title": "Test Title",
	}
	result, err := handleGhIssueCreate(context.Background(), args)

	if err != nil {
		t.Fatalf("handleGhIssueCreate returned an unexpected error: %v", err)
	}
	if !result["is_error"].(bool) {
		t.Errorf("Expected is_error to be true, got false")
	}
	if !strings.HasPrefix(result["output"].(string), "gh issue create failed:") {
		t.Errorf("Expected error message to start with 'gh issue create failed:', got '%s'", result["output"])
	}
	// Check if the simulated output is part of the actual output
	if !strings.Contains(result["output"].(string), simulatedOutput) {
		t.Errorf("Expected error message to contain '%s', got '%s'", simulatedOutput, result["output"])
	}
}
