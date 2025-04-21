package makasero

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

type FunctionHandler func(ctx context.Context, args map[string]any) (map[string]any, error)

type FunctionDefinition struct {
	Declaration *genai.FunctionDeclaration
	Handler     FunctionHandler
}

var builtinFunctions = map[string]FunctionDefinition{
	"git_add": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "git_add",
			Description: "git add を実行します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"path_to_add": {
						Type:        genai.TypeString,
						Description: "git add するファイルまたはディレクトリのパス",
					},
				},
				Required: []string{"path_to_add"},
			},
		},
		Handler: handleGitAdd,
	},
	"git_commit": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "git_commit",
			Description: "git commit を実行します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"commit_message": {
						Type:        genai.TypeString,
						Description: "git commit のメッセージ",
					},
				},
				Required: []string{"commit_message"},
			},
		},
		Handler: handleGitCommit,
	},
	"git_status": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "git_status",
			Description: "git status を実行します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"path_to_status": {
						Type:        genai.TypeString,
						Description: "git status を実行するパス",
					},
				},
				Required: []string{"path_to_status"},
			},
		},
		Handler: handleGitStatus,
	},
	"git_diff": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "git_diff",
			Description: "git diff を実行します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"path_to_diff": {
						Type:        genai.TypeString,
						Description: "git diff を実行するパス",
					},
					"staged": {
						Type:        genai.TypeBoolean,
						Description: "ステージングエリアの変更を表示するかどうか",
					},
				},
				Required: []string{"path_to_diff"},
			},
		},
		Handler: handleGitDiff,
	},
	"complete": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "complete",
			Description: "タスク完了を報告します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"message": {
						Type:        genai.TypeString,
						Description: "完了メッセージ",
					},
				},
				Required: []string{"message"},
			},
		},
		Handler: handleComplete,
	},
	"askQuestion": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "askQuestion",
			Description: "ユーザーに質問を投げかけます。タスクの遂行のためにさらに情報が必要である場合にこの関数を呼び出します。",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"question": {
						Type:        genai.TypeString,
						Description: "ユーザーへの質問内容",
					},
					"options": {
						Type:        genai.TypeArray,
						Description: "選択肢（オプション）",
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
				Required: []string{"question"},
			},
		},
		Handler: handleAskQuestion,
	},
}

func handleGitAdd(ctx context.Context, args map[string]any) (map[string]any, error) {
	pathToAdd, ok := args["path_to_add"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "path_to_add is required",
		}, nil
	}

	cmd := exec.Command("git", "add", pathToAdd)
	output, err := cmd.Output()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("git add failed: %v", err),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}

func handleGitCommit(ctx context.Context, args map[string]any) (map[string]any, error) {
	commitMessage, ok := args["commit_message"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "commit_message is required",
		}, nil
	}

	cmd := exec.Command("git", "commit", "-m", commitMessage)
	output, err := cmd.Output()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("git commit failed: %v", err),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}

func handleGitStatus(ctx context.Context, args map[string]any) (map[string]any, error) {
	pathToStatus, ok := args["path_to_status"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "path_to_status is required",
		}, nil
	}

	cmd := exec.Command("git", "status", "--short", "--", pathToStatus)
	output, err := cmd.Output()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("git status failed: %v", err),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}

func handleGitDiff(ctx context.Context, args map[string]any) (map[string]any, error) {
	pathToDiff, ok := args["path_to_diff"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "path_to_diff is required",
		}, nil
	}

	var cmd *exec.Cmd
	if staged, ok := args["staged"].(bool); ok && staged {
		cmd = exec.Command("git", "diff", "--staged", "--", pathToDiff)
	} else {
		cmd = exec.Command("git", "diff", "--", pathToDiff)
	}

	output, err := cmd.Output()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("git diff failed: %v", err),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}

func handleComplete(ctx context.Context, args map[string]any) (map[string]any, error) {
	fmt.Printf("🤖 Task completed!:\n%v\n", strings.TrimSpace(args["message"].(string)))
	return nil, nil
}

func handleAskQuestion(ctx context.Context, args map[string]any) (map[string]any, error) {
	fmt.Printf("🤖 Question:\n%v\n", strings.TrimSpace(args["question"].(string)))
	fmt.Printf("🤖 Options:\n")
	options, ok := args["options"].([]any)
	if !ok {
		// empty options is allowed
		return nil, nil
	}
	for _, option := range options {
		fmt.Printf("  %v\n", option.(string))
	}
	return nil, nil
}
