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
			checkPoint: "---COMMAND---",
		},
		{
			name:       "必須セクション：EXPLANATION",
			checkPoint: "---EXPLANATION---",
		},
		{
			name:       "必須セクション：VALIDATION",
			checkPoint: "---VALIDATION---",
		},
		{
			name:       "必須セクション：END",
			checkPoint: "---END---",
		},
		// 基本要件
		{
			name:       "要件：Linuxコマンド",
			checkPoint: "findやgrepなどのLinuxコマンドを使用",
		},
		{
			name:       "要件：シェル機能",
			checkPoint: "パイプやリダイレクトなどのシェル機能を活用",
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
			checkPoint: "指定された作業ディレクトリ内でのみ操作可能",
		},
		{
			name:       "セキュリティ：破壊的操作",
			checkPoint: "rm, mv, cp等の破壊的な操作は禁止",
		},
		{
			name:       "セキュリティ：sudo制限",
			checkPoint: "sudoの使用は禁止",
		},
		// システム情報
		{
			name:       "システム情報：OS",
			checkPoint: "OSの種類とバージョン",
		},
		{
			name:       "システム情報：シェル",
			checkPoint: "シェルの種類",
		},
		{
			name:       "システム情報：作業ディレクトリ",
			checkPoint: "現在の作業ディレクトリ",
		},
		// エラーハンドリング
		{
			name:       "エラー：構文",
			checkPoint: "コマンドの構文エラー",
		},
		{
			name:       "エラー：権限",
			checkPoint: "権限エラー",
		},
		{
			name:       "エラー：リソース",
			checkPoint: "リソース制限エラー",
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
		name   string
		marker CommandMarkers
	}{
		{
			name: "正常系：デフォルトマーカー",
			marker: CommandMarkers{
				Command:     "---COMMAND---",
				Explanation: "---EXPLANATION---",
				Validation:  "---VALIDATION---",
				End:         "---END---",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.marker, DefaultCommandMarkers)
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
				Success:  true,
				Output:   "test output",
				Error:    "",
				ExitCode: 0,
				Duration: 100,
			},
			wantErr: false,
		},
		{
			name: "正常系：エラー",
			result: CommandResult{
				Success:  false,
				Output:   "",
				Error:    "command not found",
				ExitCode: 1,
				Duration: 50,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, !tt.wantErr, tt.result.Success)
			if tt.wantErr {
				assert.NotEmpty(t, tt.result.Error)
				assert.NotEqual(t, 0, tt.result.ExitCode)
			} else {
				assert.Empty(t, tt.result.Error)
				assert.Equal(t, 0, tt.result.ExitCode)
			}
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
				Details: "The command 'rm' is not allowed",
			},
			check: func(t *testing.T, err ValidationError) {
				assert.NotEmpty(t, err.Code)
				assert.NotEmpty(t, err.Message)
				assert.NotEmpty(t, err.Details)
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
				WorkDir:        "/tmp",
				AllowedCmds:    []string{"ls", "grep", "find"},
				BlockedCmds:    []string{"rm", "mv", "cp"},
				MaxCPUPercent:  50,
				MaxMemoryMB:    1024,
				TimeoutSeconds: 30,
			},
			check: func(t *testing.T, c SecurityConstraints) {
				assert.NotEmpty(t, c.WorkDir)
				assert.NotEmpty(t, c.AllowedCmds)
				assert.NotEmpty(t, c.BlockedCmds)
				assert.Greater(t, c.MaxCPUPercent, 0)
				assert.Greater(t, c.MaxMemoryMB, 0)
				assert.Greater(t, c.TimeoutSeconds, 0)
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
	assert.True(t, strings.Contains(CommandSystemPrompt, DefaultCommandMarkers.Command))
	assert.True(t, strings.Contains(CommandSystemPrompt, DefaultCommandMarkers.Explanation))
	assert.True(t, strings.Contains(CommandSystemPrompt, DefaultCommandMarkers.Validation))
	assert.True(t, strings.Contains(CommandSystemPrompt, DefaultCommandMarkers.End))

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
