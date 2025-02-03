package prompts

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandSystemPrompt(t *testing.T) {
	tests := []struct {
		name       string
		checkPoint string
	}{
		// 必須セクション
		{
			name:       "必須セクション：COMMAND",
			checkPoint: CommandMarkerStart,
		},
		{
			name:       "必須セクション：EXPLANATION",
			checkPoint: CommandMarkerExplanation,
		},
		{
			name:       "必須セクション：VALIDATION",
			checkPoint: CommandMarkerValidation,
		},
		{
			name:       "必須セクション：END",
			checkPoint: CommandMarkerEnd,
		},
		// 基本要件
		{
			name:       "要件：Linuxコマンド",
			checkPoint: "findやgrepなどの検索系コマンドを使用",
		},
		{
			name:       "要件：シェル機能",
			checkPoint: "パイプやリダイレクトは必要に応じて使用",
		},
		{
			name:       "要件：読み取り専用",
			checkPoint: "読み取り専用の操作のみ許可",
		},
		{
			name:       "要件：表示形式",
			checkPoint: "結果は人間が理解しやすい形式で表示",
		},
		// セキュリティ制約
		{
			name:       "セキュリティ：ディレクトリ制限",
			checkPoint: "作業ディレクトリ内でのみ操作",
		},
		{
			name:       "セキュリティ：破壊的操作",
			checkPoint: "破壊的な操作は禁止",
		},
		{
			name:       "セキュリティ：特権コマンド",
			checkPoint: "特権コマンドの使用は禁止",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, CommandSystemPrompt, tt.checkPoint)
		})
	}
}

func TestCommandMarkers(t *testing.T) {
	tests := []struct {
		name  string
		start string
		exp   string
		val   string
		end   string
	}{
		{
			name:  "正常系：デフォルトマーカー",
			start: CommandMarkerStart,
			exp:   CommandMarkerExplanation,
			val:   CommandMarkerValidation,
			end:   CommandMarkerEnd,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, CommandMarkerStart, tt.start)
			assert.Equal(t, CommandMarkerExplanation, tt.exp)
			assert.Equal(t, CommandMarkerValidation, tt.val)
			assert.Equal(t, CommandMarkerEnd, tt.end)
		})
	}
}

func TestCommandResult(t *testing.T) {
	tests := []struct {
		name    string
		result  CommandResult
		wantErr bool
	}{
		{
			name: "正常系：成功",
			result: CommandResult{
				Command:     "find . -name \"*.txt\"",
				Explanation: "テキストファイルを検索します",
				Validation:  "読み取り専用の安全な操作です",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.result.Command)
			assert.NotEmpty(t, tt.result.Explanation)
			assert.NotEmpty(t, tt.result.Validation)
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name  string
		err   ValidationError
		check func(*testing.T, ValidationError)
	}{
		{
			name: "正常系：必須フィールド",
			err: ValidationError{
				Code:    "INVALID_COMMAND",
				Message: "Invalid command",
			},
			check: func(t *testing.T, err ValidationError) {
				assert.NotEmpty(t, err.Code)
				assert.NotEmpty(t, err.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.err)
		})
	}
}

func TestSecurityConstraints(t *testing.T) {
	tests := []struct {
		name        string
		constraints SecurityConstraints
		check       func(*testing.T, SecurityConstraints)
	}{
		{
			name: "正常系：デフォルト値",
			constraints: SecurityConstraints{
				WorkDir:         "/tmp",
				AllowedCommands: []string{"ls", "grep", "find"},
				ResourceLimits: map[string]string{
					"cpu":    "50%",
					"memory": "1024MB",
					"time":   "30s",
				},
			},
			check: func(t *testing.T, c SecurityConstraints) {
				assert.NotEmpty(t, c.WorkDir)
				assert.NotEmpty(t, c.AllowedCommands)
				assert.NotEmpty(t, c.ResourceLimits)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.constraints)
		})
	}
}

func TestSystemInfo(t *testing.T) {
	tests := []struct {
		name  string
		info  SystemInfo
		check func(*testing.T, SystemInfo)
	}{
		{
			name: "正常系：必須フィールド",
			info: SystemInfo{
				OS:      "Linux",
				Shell:   "/bin/bash",
				WorkDir: "/home/user",
				Environment: map[string]string{
					"PATH": "/usr/bin:/bin",
					"HOME": "/home/user",
				},
			},
			check: func(t *testing.T, info SystemInfo) {
				assert.NotEmpty(t, info.OS)
				assert.NotEmpty(t, info.Shell)
				assert.NotEmpty(t, info.WorkDir)
				assert.NotEmpty(t, info.Environment)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.info)
		})
	}
}

func TestCommandSystemPrompt_Integration(t *testing.T) {
	// プロンプトにマーカーが含まれていることを確認
	assert.True(t, strings.Contains(CommandSystemPrompt, CommandMarkerStart))
	assert.True(t, strings.Contains(CommandSystemPrompt, CommandMarkerExplanation))
	assert.True(t, strings.Contains(CommandSystemPrompt, CommandMarkerValidation))
	assert.True(t, strings.Contains(CommandSystemPrompt, CommandMarkerEnd))

	// プロンプトの形式が正しいことを確認
	lines := strings.Split(CommandSystemPrompt, "\n")
	var hasRequirements, hasSecurity, hasSystemInfo, hasErrorHandling bool
	for _, line := range lines {
		if strings.Contains(line, "要件：") {
			hasRequirements = true
		}
		if strings.Contains(line, "セキュリティ制約：") {
			hasSecurity = true
		}
		if strings.Contains(line, "システム情報の考慮：") {
			hasSystemInfo = true
		}
		if strings.Contains(line, "エラーハンドリング：") {
			hasErrorHandling = true
		}
	}
	assert.True(t, hasRequirements, "プロンプトに要件セクションが含まれていること")
	assert.True(t, hasSecurity, "プロンプトにセキュリティ制約セクションが含まれていること")
	assert.True(t, hasSystemInfo, "プロンプトにシステム情報セクションが含まれていること")
	assert.True(t, hasErrorHandling, "プロンプトにエラーハンドリングセクションが含まれていること")
}
