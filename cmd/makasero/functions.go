package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

type FunctionHandler func(ctx context.Context, args map[string]any) (map[string]any, error)

type FunctionDefinition struct {
	Declaration *genai.FunctionDeclaration
	Handler     FunctionHandler
}

var functions = map[string]FunctionDefinition{
	/*
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
	*/
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
	/*
			"readFile": {
				Declaration: &genai.FunctionDeclaration{
					Name:        "readFile",
					Description: "指定されたファイルの内容を読み取ります",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {
								Type:        genai.TypeString,
								Description: "読み取るファイルのパス",
							},
							"startLine": {
								Type:        genai.TypeInteger,
								Description: "読み取り開始行（1から始まる、省略時は1行目から）",
							},
							"endLine": {
								Type:        genai.TypeInteger,
								Description: "読み取り終了行（1から始まる、省略時は最終行まで）",
							},
						},
						Required: []string{"path"},
					},
				},
				Handler: handleReadFile,
			},
			"writeFile": {
				Declaration: &genai.FunctionDeclaration{
					Name:        "writeFile",
					Description: "指定されたファイルに内容を書き込みます",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {
								Type:        genai.TypeString,
								Description: "書き込むファイルのパス",
							},
							"content": {
								Type:        genai.TypeString,
								Description: "書き込む内容",
							},
							"append": {
								Type:        genai.TypeBoolean,
								Description: "追記モードかどうか（デフォルト: false）",
							},
							"startLine": {
								Type:        genai.TypeInteger,
								Description: "書き込み開始行（1から始まる、省略時はファイルの末尾）",
							},
						},
						Required: []string{"path", "content"},
					},
				},
				Handler: handleWriteFile,
			},
			"createPath": {
				Declaration: &genai.FunctionDeclaration{
					Name:        "createPath",
					Description: "指定されたパスを作成します（ディレクトリやファイル）",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {
								Type:        genai.TypeString,
								Description: "作成するパス",
							},
							"type": {
								Type:        genai.TypeString,
								Description: "作成するタイプ（'file' または 'directory'）",
								Enum:        []string{"file", "directory"},
							},
							"parents": {
								Type:        genai.TypeBoolean,
								Description: "親ディレクトリも作成するかどうか（デフォルト: false）",
							},
							"mode": {
								Type:        genai.TypeInteger,
								Description: "作成するファイル/ディレクトリのパーミッション（8進数、デフォルト: 0755）",
							},
						},
						Required: []string{"path", "type"},
					},
				},
				Handler: handleCreatePath,
			},
			"apply_patch": {
				Declaration: &genai.FunctionDeclaration{
					Name:        "apply_patch",
					Description: "指定されたパッチを適用します",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"patch": {
								Type:        genai.TypeString,
								Description: "適用するパッチの内容",
							},
							"path": {
								Type:        genai.TypeString,
								Description: "パッチを適用するファイルのパス",
							},
							"reverse": {
								Type:        genai.TypeBoolean,
								Description: "パッチを逆に適用するかどうか（デフォルト: false）",
							},
							"strip": {
								Type:        genai.TypeInteger,
								Description: "パスから取り除くディレクトリの数（デフォルト: 0）",
							},
						},
						Required: []string{"patch", "path"},
					},
				},
				Handler: handleApplyPatch,
			},

		"getGitHubPRDiff": {
			Declaration: &genai.FunctionDeclaration{
				Name:        "getGitHubPRDiff",
				Description: "GitHubのプルリクエストのdiffを取得します",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"repository": {
							Type:        genai.TypeString,
							Description: "リポジトリ名（例: pankona/makasero）",
						},
						"prNumber": {
							Type:        genai.TypeInteger,
							Description: "プルリクエスト番号",
						},
					},
					Required: []string{"repository", "prNumber"},
				},
			},
			Handler: handleGetGitHubPRDiff,
		},
	*/
	"askQuestion": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "askQuestion",
			Description: "ユーザーに質問を投げかけます",
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
	if args["startLine"] != nil {
		if sl, ok := args["startLine"].(float64); ok {
			startLine = int(sl)
		}
	}

	endLine := -1 // -1 means read until the end
	if args["endLine"] != nil {
		if el, ok := args["endLine"].(float64); ok {
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
			"error":   fmt.Sprintf("startLine %d is out of range (1-%d)", startLine, len(lines)),
		}, nil
	}

	if endLine == -1 {
		endLine = len(lines)
	} else if endLine < startLine || endLine > len(lines) {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("endLine %d is out of range (%d-%d)", endLine, startLine, len(lines)),
		}, nil
	}

	selectedLines := lines[startLine-1 : endLine]
	return map[string]any{
		"success":    true,
		"content":    strings.Join(selectedLines, "\n"),
		"path":       path,
		"startLine":  startLine,
		"endLine":    endLine,
		"totalLines": len(lines),
	}, nil
}

func handleWriteFile(ctx context.Context, args map[string]any) (map[string]any, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	appendMode := false
	if args["append"] != nil {
		appendMode = args["append"].(bool)
	}

	startLine := -1 // -1 means append to the end
	if args["startLine"] != nil {
		if sl, ok := args["startLine"].(float64); ok {
			startLine = int(sl)
		}
	}

	var file *os.File
	var err error

	if appendMode {
		file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	} else if startLine == -1 {
		file, err = os.Create(path)
	} else {
		// 既存のファイルを読み込んで、指定行に挿入
		existingContent, err := os.ReadFile(path)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, nil
		}

		lines := strings.Split(string(existingContent), "\n")
		if startLine < 1 || startLine > len(lines) {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("startLine %d is out of range (1-%d)", startLine, len(lines)),
			}, nil
		}

		// 新しい内容を挿入
		contentLines := strings.Split(content, "\n")
		newLines := make([]string, 0, len(lines)+len(contentLines))
		newLines = append(newLines, lines[:startLine-1]...)
		newLines = append(newLines, contentLines...)
		newLines = append(newLines, lines[startLine-1:]...)
		newContent := strings.Join(newLines, "\n")

		file, err = os.Create(path)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, nil
		}
		_, err = file.WriteString(newContent)
		if err != nil {
			file.Close()
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, nil
		}
		file.Close()
		return map[string]any{
			"success": true,
			"path":    path,
			"lines":   len(newLines),
		}, nil
	}

	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"path":    path,
	}, nil
}

func handleCreatePath(ctx context.Context, args map[string]any) (map[string]any, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	pathType, ok := args["type"].(string)
	if !ok {
		return nil, fmt.Errorf("type is required")
	}

	parents := false
	if args["parents"] != nil {
		parents = args["parents"].(bool)
	}

	mode := os.FileMode(0755)
	if args["mode"] != nil {
		switch m := args["mode"].(type) {
		case float64:
			mode = os.FileMode(m)
		case string:
			// 8進数の文字列を数値に変換
			var m64 int64
			if _, err := fmt.Sscanf(m, "%o", &m64); err == nil {
				mode = os.FileMode(m64)
			}
		}
	}

	var err error
	if pathType == "directory" {
		if parents {
			err = os.MkdirAll(path, mode)
		} else {
			err = os.Mkdir(path, mode)
		}
	} else if pathType == "file" {
		if parents {
			dir := filepath.Dir(path)
			if err = os.MkdirAll(dir, mode); err != nil {
				return map[string]any{
					"success": false,
					"error":   err.Error(),
				}, nil
			}
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, mode)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   err.Error(),
			}, nil
		}
		file.Close()
	} else {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("invalid type: %s", pathType),
		}, nil
	}

	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"path":    path,
		"type":    pathType,
	}, nil
}

func handleApplyPatch(ctx context.Context, args map[string]any) (map[string]any, error) {
	patch, ok := args["patch"].(string)
	if !ok {
		return nil, fmt.Errorf("patch is required")
	}

	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	reverse := false
	if args["reverse"] != nil {
		reverse = args["reverse"].(bool)
	}

	strip := 0
	if args["strip"] != nil {
		if s, ok := args["strip"].(float64); ok {
			strip = int(s)
		}
	}

	cmd := exec.Command("patch")
	if reverse {
		cmd.Args = append(cmd.Args, "-R")
	}
	if strip > 0 {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-p%d", strip))
	}
	cmd.Args = append(cmd.Args, path)

	cmd.Stdin = strings.NewReader(patch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
			"output":  string(output),
		}, nil
	}

	return map[string]any{
		"success": true,
		"path":    path,
		"output":  string(output),
	}, nil
}

func handleGetGitHubPRDiff(ctx context.Context, args map[string]any) (map[string]any, error) {
	repository, ok := args["repository"].(string)
	if !ok {
		return nil, fmt.Errorf("repository is required")
	}

	prNumber, ok := args["prNumber"].(float64)
	if !ok {
		return nil, fmt.Errorf("prNumber is required")
	}

	cmd := exec.Command("gh", "pr", "diff", fmt.Sprintf("%d", int(prNumber)), "--repo", repository)
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

func handleAskQuestion(ctx context.Context, args map[string]any) (map[string]any, error) {
	// 呼び出されないので実装しなくてよい
	return nil, nil
}
