package models

import (
	"fmt"
	"os"
	"testing"
)

// モックAPIクライアント
type mockAPIClient struct {
	response string
	err      error
}

func (m *mockAPIClient) CreateChatCompletion(messages []ChatMessage) (string, error) {
	return m.response, m.err
}

func TestTestCommandAnalyzer_AnalyzePrompt(t *testing.T) {
	tests := []struct {
		name            string
		prompt          string
		mockResponse    string
		mockErr         error
		wantCommand     string
		wantExplanation string
		wantOk          bool
	}{
		{
			name:   "テストコマンドの提案が成功する場合",
			prompt: "テストを実行して",
			mockResponse: `---COMMAND---
go test ./...
---EXPLANATION---
すべてのパッケージのテストを実行します
---END---`,
			mockErr:         nil,
			wantCommand:     "go test ./...",
			wantExplanation: "すべてのパッケージのテストを実行します",
			wantOk:          true,
		},
		{
			name:         "テストに関係ない入力の場合",
			prompt:       "ファイルを作成して",
			mockResponse: "NOT_TEST",
			mockErr:      nil,
			wantOk:       false,
		},
		{
			name:    "APIエラーの場合",
			prompt:  "テストを実行して",
			mockErr: fmt.Errorf("API error"),
			wantOk:  false,
		},
		{
			name:         "不正なレスポンス形式の場合",
			prompt:       "テストを実行して",
			mockResponse: "invalid format",
			mockErr:      nil,
			wantOk:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAPIClient{
				response: tt.mockResponse,
				err:      tt.mockErr,
			}
			analyzer := NewTestCommandAnalyzer()
			proposal, ok := analyzer.AnalyzePrompt(tt.prompt, client)

			if ok != tt.wantOk {
				t.Errorf("AnalyzePrompt() ok = %v, want %v", ok, tt.wantOk)
				return
			}

			if !ok {
				if proposal != nil {
					t.Errorf("AnalyzePrompt() proposal = %v, want nil", proposal)
				}
				return
			}

			if proposal.Command != tt.wantCommand {
				t.Errorf("AnalyzePrompt() command = %v, want %v", proposal.Command, tt.wantCommand)
			}

			if proposal.Explanation != tt.wantExplanation {
				t.Errorf("AnalyzePrompt() explanation = %v, want %v", proposal.Explanation, tt.wantExplanation)
			}

			if proposal.Type != "test" {
				t.Errorf("AnalyzePrompt() type = %v, want test", proposal.Type)
			}
		})
	}
}

func TestParseCommandResponse(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		wantCommand     string
		wantExplanation string
		wantErr         bool
	}{
		{
			name: "正常なレスポンスの場合",
			response: `---COMMAND---
go test ./...
---EXPLANATION---
すべてのパッケージのテストを実行します
---END---`,
			wantCommand:     "go test ./...",
			wantExplanation: "すべてのパッケージのテストを実行します",
			wantErr:         false,
		},
		{
			name:     "不正なフォーマットの場合",
			response: "invalid format",
			wantErr:  true,
		},
		{
			name: "コマンドが空の場合",
			response: `---COMMAND---
---EXPLANATION---
説明
---END---`,
			wantCommand:     "",
			wantExplanation: "説明",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, explanation, err := parseCommandResponse(tt.response)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseCommandResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if command != tt.wantCommand {
				t.Errorf("parseCommandResponse() command = %v, want %v", command, tt.wantCommand)
			}

			if explanation != tt.wantExplanation {
				t.Errorf("parseCommandResponse() explanation = %v, want %v", explanation, tt.wantExplanation)
			}
		})
	}
}

// モックコマンド実行
type mockCommandExecutor struct {
	executed    string
	outputs     []string
	errs        []error
	execCounter int
}

func (m *mockCommandExecutor) Execute(command string) (string, error) {
	m.executed = command
	if m.execCounter < len(m.outputs) && m.execCounter < len(m.errs) {
		output := m.outputs[m.execCounter]
		err := m.errs[m.execCounter]
		m.execCounter++
		return output, err
	}
	return "", nil
}

func TestCommandRunner_RunWithApproval(t *testing.T) {
	// 標準入力をモック
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	tests := []struct {
		name          string
		proposal      *CommandProposal
		inputs        []string // 複数の入力を配列で管理
		execOutputs   []string // 複数の実行結果を配列で管理
		execErrs      []error  // 複数のエラーを配列で管理
		mockResponse  string
		mockAPIErr    error
		wantErr       bool
		wantExec      string
		wantExecCount int
	}{
		{
			name: "コマンドが承認され正常に実行される",
			proposal: &CommandProposal{
				Command:     "echo 'test'",
				Explanation: "テストコマンド",
				Type:        "test",
			},
			inputs:      []string{"y\n"},
			execOutputs: []string{"test\n"},
			execErrs:    []error{nil},
			wantErr:     false,
			wantExec:    "echo 'test'",
		},
		{
			name: "コマンドが拒否される",
			proposal: &CommandProposal{
				Command:     "rm -rf /",
				Explanation: "危険なコマンド",
				Type:        "test",
			},
			inputs:  []string{"n\n"},
			wantErr: true,
		},
		{
			name:     "提案がnilの場合",
			proposal: nil,
			wantErr:  true,
		},
		{
			name: "コマンド実行に失敗し修正案が提案される",
			proposal: &CommandProposal{
				Command:     "go test",
				Explanation: "テストの実行",
				Type:        "test",
			},
			inputs:      []string{"y\n", "y\n"},
			execOutputs: []string{"package not found", "all tests passed"},
			execErrs:    []error{fmt.Errorf("exit status 1"), nil},
			mockResponse: `---COMMAND---
go test ./...
---EXPLANATION---
パッケージパスを修正しました
---END---`,
			wantExec: "go test ./...",
		},
		{
			name: "コマンド実行に失敗し修正が不要と判断される",
			proposal: &CommandProposal{
				Command:     "invalid",
				Explanation: "不正なコマンド",
				Type:        "test",
			},
			inputs:       []string{"y\n"},
			execOutputs:  []string{"command not found"},
			execErrs:     []error{fmt.Errorf("exit status 127")},
			mockResponse: "NO_FIX_NEEDED",
			wantErr:      true,
			wantExec:     "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックExecutorの作成
			executor := &mockCommandExecutor{
				outputs: tt.execOutputs,
				errs:    tt.execErrs,
			}

			// モックAPIクライアントの作成
			client := &mockAPIClient{
				response: tt.mockResponse,
				err:      tt.mockAPIErr,
			}

			// 標準入力のモック
			if len(tt.inputs) > 0 {
				pipeReader, pipeWriter, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				os.Stdin = pipeReader

				// 入力を書き込む
				go func() {
					defer pipeWriter.Close()
					for _, input := range tt.inputs {
						pipeWriter.Write([]byte(input))
					}
				}()
			}

			runner := NewCommandRunner(executor, client)
			err := runner.RunWithApproval(tt.proposal)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunWithApproval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantExec != "" && executor.executed != tt.wantExec {
				t.Errorf("最後に実行されたコマンドが異なります。got = %v, want %v", executor.executed, tt.wantExec)
			}
		})
	}
}
