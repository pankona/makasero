package main

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

var functions = map[string]FunctionDefinition{
	"execCommand": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "execCommand",
			Description: "ターミナルコマンドを実行し、その結果を返します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"command": {
						Type:        genai.TypeString,
						Description: "実行するコマンド",
					},
					"args": {
						Type:        genai.TypeArray,
						Description: "コマンドの引数",
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
				Required: []string{"command"},
			},
		},
		Handler: handleExecCommand,
	},
	"getGitHubIssue": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "getGitHubIssue",
			Description: "GitHubのIssueの詳細を取得します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"repository": {
						Type:        genai.TypeString,
						Description: "リポジトリ名（例: pankona/makasero）",
					},
					"issueNumber": {
						Type:        genai.TypeInteger,
						Description: "Issue番号",
					},
				},
				Required: []string{"repository", "issueNumber"},
			},
		},
		Handler: handleGetGitHubIssue,
	},
	"gitStatus": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "gitStatus",
			Description: "Gitのステータスを表示します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"path": {
						Type:        genai.TypeString,
						Description: "ステータスを確認するパス（オプション）",
					},
				},
			},
		},
		Handler: handleGitStatus,
	},
	"gitAdd": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "gitAdd",
			Description: "Gitの変更をステージングします",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"paths": {
						Type:        genai.TypeArray,
						Description: "ステージングするファイルのパス（オプション）",
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"all": {
						Type:        genai.TypeBoolean,
						Description: "全ての変更をステージングするかどうか（デフォルト: false）",
					},
				},
			},
		},
		Handler: handleGitAdd,
	},
	"gitCommit": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "gitCommit",
			Description: "Gitの変更をコミットします",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"message": {
						Type:        genai.TypeString,
						Description: "コミットメッセージ",
					},
					"type": {
						Type:        genai.TypeString,
						Description: "コミットの種類（例: feat, fix, refactor など）",
					},
					"scope": {
						Type:        genai.TypeString,
						Description: "コミットのスコープ（例: パッケージ名）",
					},
					"description": {
						Type:        genai.TypeString,
						Description: "コミットの詳細な説明",
					},
				},
				Required: []string{"message"},
			},
		},
		Handler: handleGitCommit,
	},
	"gitDiff": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "gitDiff",
			Description: "Gitの変更差分を表示します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"path": {
						Type:        genai.TypeString,
						Description: "差分を確認するパス（オプション）",
					},
					"staged": {
						Type:        genai.TypeBoolean,
						Description: "ステージングされた変更の差分を表示するかどうか（デフォルト: false）",
					},
				},
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
	"read_file": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "read_file",
			Description: "指定されたファイルの内容を読み取ります",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"path": {
						Type:        genai.TypeString,
						Description: "読み取るファイルのパス",
					},
					"start_line": {
						Type:        genai.TypeInteger,
						Description: "読み取り開始行（1から始まる、省略時は1行目から）",
					},
					"end_line": {
						Type:        genai.TypeInteger,
						Description: "読み取り終了行（1から始まる、省略時は最終行まで）",
					},
				},
				Required: []string{"path"},
			},
		},
		Handler: handleReadFile,
	},
}

func handleExecCommand(ctx context.Context, args map[string]any) (map[string]any, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command is required")
	}

	var cmdArgs []string
	if args["args"] != nil {
		argsList, ok := args["args"].([]any)
		if !ok {
			return nil, fmt.Errorf("args must be an array")
		}
		for _, arg := range argsList {
			if str, ok := arg.(string); ok {
				cmdArgs = append(cmdArgs, str)
			}
		}
	}

	cmd := exec.Command(command, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"output":  string(output),
	}, nil
}

func handleGetGitHubIssue(ctx context.Context, args map[string]any) (map[string]any, error) {
	repository, ok := args["repository"].(string)
	if !ok {
		return nil, fmt.Errorf("repository is required")
	}

	issueNumber, ok := args["issueNumber"].(float64)
	if !ok {
		return nil, fmt.Errorf("issueNumber is required")
	}

	cmd := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", int(issueNumber)), "--repo", repository, "--json", "title,body,state,labels,createdAt,updatedAt,assignees,milestone,comments")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"raw":     string(output),
	}, nil
}

func handleGitStatus(ctx context.Context, args map[string]any) (map[string]any, error) {
	var path string
	if args["path"] != nil {
		path, _ = args["path"].(string)
	}

	cmd := exec.Command("git", "status")
	if path != "" {
		cmd.Args = append(cmd.Args, path)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"status":  string(output),
	}, nil
}

func handleGitAdd(ctx context.Context, args map[string]any) (map[string]any, error) {
	var paths []string
	if args["paths"] != nil {
		pathsList, ok := args["paths"].([]any)
		if !ok {
			return nil, fmt.Errorf("paths must be an array")
		}
		for _, path := range pathsList {
			if str, ok := path.(string); ok {
				paths = append(paths, str)
			}
		}
	}

	all := false
	if args["all"] != nil {
		all, _ = args["all"].(bool)
	}

	cmd := exec.Command("git", "add")
	if all {
		cmd.Args = append(cmd.Args, "--all")
	} else if len(paths) > 0 {
		cmd.Args = append(cmd.Args, paths...)
	} else {
		cmd.Args = append(cmd.Args, ".")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"output":  string(output),
	}, nil
}

func handleGitCommit(ctx context.Context, args map[string]any) (map[string]any, error) {
	message := ""
	if args["message"] != nil {
		message = args["message"].(string)
	}
	commitType := ""
	if args["type"] != nil {
		commitType = args["type"].(string)
	}
	scope := ""
	if args["scope"] != nil {
		scope = args["scope"].(string)
	}
	description := ""
	if args["description"] != nil {
		description = args["description"].(string)
	}

	cmd := exec.Command("git", "commit", "-m", message)
	if commitType != "" {
		cmd.Args = append(cmd.Args, "-m", fmt.Sprintf("%s: %s", commitType, message))
	}
	if scope != "" {
		cmd.Args = append(cmd.Args, "-m", fmt.Sprintf("(%s): %s", scope, description))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"output":  string(output),
	}, nil
}

func handleGitDiff(ctx context.Context, args map[string]any) (map[string]any, error) {
	var path string
	if args["path"] != nil {
		path = args["path"].(string)
	}
	var staged bool
	if args["staged"] != nil {
		staged = args["staged"].(bool)
	}

	cmd := exec.Command("git", "diff")
	if path != "" {
		cmd.Args = append(cmd.Args, path)
	}
	if staged {
		cmd.Args = append(cmd.Args, "--cached")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"diff":    string(output),
	}, nil
}

func handleComplete(ctx context.Context, args map[string]any) (map[string]any, error) {
	// 呼び出されないので実装しなくてよい
	return nil, nil
}

func handleReadFile(ctx context.Context, args map[string]any) (map[string]any, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	startLine := 1
	if args["start_line"] != nil {
		if sl, ok := args["start_line"].(float64); ok {
			startLine = int(sl)
		}
	}

	endLine := -1 // -1 means read until the end
	if args["end_line"] != nil {
		if el, ok := args["end_line"].(float64); ok {
			endLine = int(el)
		}
	}

	// catコマンドでファイルを読み取る
	cmd := exec.Command("cat", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	lines := strings.Split(string(output), "\n")
	// 最後の空行を削除
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if startLine < 1 || startLine > len(lines) {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("start_line %d is out of range (1-%d)", startLine, len(lines)),
		}, nil
	}

	if endLine == -1 {
		endLine = len(lines)
	} else if endLine < startLine || endLine > len(lines) {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("end_line %d is out of range (%d-%d)", endLine, startLine, len(lines)),
		}, nil
	}

	selectedLines := lines[startLine-1 : endLine]
	return map[string]any{
		"success":     true,
		"content":     strings.Join(selectedLines, "\n"),
		"path":        path,
		"start_line":  startLine,
		"end_line":    endLine,
		"total_lines": len(lines),
	}, nil
}
