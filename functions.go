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
	"ask_question": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "ask_question",
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
	"gh_issue_view": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "gh_issue_view",
			Description: "gh issue view コマンドを使って、指定された番号の GitHub issue を表示します。",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"issue_number": {
						Type:        genai.TypeNumber,
						Description: "表示する GitHub issue の番号",
					},
					"repo": {
						Type:        genai.TypeString,
						Description: "リポジトリ名 (例: owner/repo)。指定がない場合は現在のリポジトリとみなされます。",
					},
				},
				Required: []string{"issue_number"},
			},
		},
		Handler: handleGhIssueView,
	},
	"gh_issue_create": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "gh_issue_create",
			Description: "gh issue create コマンドを使って、GitHub Issue を作成します。",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "Issue のタイトル",
					},
					"body": {
						Type:        genai.TypeString,
						Description: "Issue の本文",
					},
					"labels": {
						Type:        genai.TypeArray,
						Description: "付与するラベルの配列 (例: [\"bug\", \"critical\"])",
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"repo": {
						Type:        genai.TypeString,
						Description: "リポジトリ名 (例: owner/repo)。指定がない場合は現在のリポジトリとみなされます。",
					},
				},
				Required: []string{"title", "body"},
			},
		},
		Handler: handleGhIssueCreate,
	},
	"create_makasero_enhancement_issue": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "create_makasero_enhancement_issue",
			Description: "makasero 自身の改善案を GitHub Issue として起票します。issue は pankona/makasero リポジトリの issue として起票され、自動的に 'enhancement' ラベルが付与されます。",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "Issue のタイトル",
					},
					"body": {
						Type:        genai.TypeString,
						Description: "Issue の本文",
					},
				},
				Required: []string{"title", "body"},
			},
		},
		Handler: handleCreateEnhancementIssue,
	},
	"gh_pr_view": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "gh_pr_view",
			Description: "gh pr view コマンドを使って、指定された番号の GitHub Pull Request を表示します。差分の確認やレビューに役立ちます。",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"pr_number": {
						Type:        genai.TypeNumber,
						Description: "表示する GitHub Pull Request の番号",
					},
					"repo": {
						Type:        genai.TypeString,
						Description: "リポジトリ名 (例: owner/repo)。指定がない場合は現在のリポジトリとみなされます。",
					},
					"diff": {
						Type:        genai.TypeBoolean,
						Description: "差分を表示するかどうか (--diff オプション)",
					},
				},
				Required: []string{"pr_number"},
			},
		},
		Handler: handleGhPrView,
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

func handleGhIssueView(ctx context.Context, args map[string]any) (map[string]any, error) {
	issueNumber, ok := args["issue_number"].(float64)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "issue_number is required and must be a number",
		}, nil
	}

	repo, _ := args["repo"].(string)

	var cmd *exec.Cmd
	if repo != "" {
		cmd = exec.Command("gh", "issue", "view", fmt.Sprintf("%.0f", issueNumber), "--repo", repo)
	} else {
		cmd = exec.Command("gh", "issue", "view", fmt.Sprintf("%.0f", issueNumber))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("gh issue view failed: %v\nOutput: %s", err, string(output)),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}

func handleGhIssueCreate(ctx context.Context, args map[string]any) (map[string]any, error) {
	title, ok := args["title"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "title is required",
		}, nil
	}

	body, ok := args["body"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "body is required",
		}, nil
	}

	repo, _ := args["repo"].(string)

	var cmdArgs []string
	cmdArgs = append(cmdArgs, "issue", "create", "--title", title, "--body", body)

	if labelsStr, ok := args["labels"].([]any); ok {
		for _, label := range labelsStr {
			if labelStr, ok := label.(string); ok {
				cmdArgs = append(cmdArgs, "--label", labelStr)
			}
		}
	}

	if repo != "" {
		cmdArgs = append(cmdArgs, "--repo", repo)
	}

	cmd := exec.Command("gh", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("gh issue create failed: %v\nOutput: %s", err, string(output)),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}

func handleCreateEnhancementIssue(ctx context.Context, args map[string]any) (map[string]any, error) {
	title, ok := args["title"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "title is required",
		}, nil
	}

	body, ok := args["body"].(string)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "body is required",
		}, nil
	}

	// repo パラメータの受付を削除し、固定値を設定
	const fixedRepo = "pankona/makasero"

	var cmdArgs []string
	// 改善提案なので、必ず enhancement ラベルを付与し、固定のリポジトリを指定する
	cmdArgs = append(cmdArgs, "issue", "create", "--title", title, "--body", body, "--label", "enhancement", "--repo", fixedRepo)

	cmd := exec.Command("gh", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("gh issue create for enhancement failed: %v\\nOutput: %s", err, string(output)),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}

func handleGhPrView(ctx context.Context, args map[string]any) (map[string]any, error) {
	prNumber, ok := args["pr_number"].(float64)
	if !ok {
		return map[string]any{
			"is_error": true,
			"output":   "pr_number is required and must be a number",
		}, nil
	}

	repo, _ := args["repo"].(string)
	diff, _ := args["diff"].(bool)

	var cmdArgs []string
	cmdArgs = append(cmdArgs, "pr", "view", fmt.Sprintf("%.0f", prNumber))

	if diff {
		cmdArgs = append(cmdArgs, "--diff")
	}

	if repo != "" {
		cmdArgs = append(cmdArgs, "--repo", repo)
	}

	cmd := exec.Command("gh", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"is_error": true,
			"output":   fmt.Sprintf("gh pr view failed: %v\nOutput: %s", err, string(output)),
		}, nil
	}

	return map[string]any{
		"is_error": false,
		"output":   string(output),
	}, nil
}
