package models

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// APIClient はLLMとの通信を行うインターフェース
type APIClient interface {
	CreateChatCompletion(messages []ChatMessage) (string, error)
}

// CommandProposal はAIが提案するコマンドを表す構造体
type CommandProposal struct {
	Command     string // 実行するコマンド
	Explanation string // コマンドの説明
	Type        string // コマンドの種類（test, git, file など）
}

// CommandAnalyzer はユーザーの入力からコマンドを分析するインターフェース
type CommandAnalyzer interface {
	AnalyzePrompt(prompt string, client APIClient) (*CommandProposal, bool)
}

// TestCommandAnalyzer はテスト関連のコマンドを分析する
type TestCommandAnalyzer struct{}

func NewTestCommandAnalyzer() *TestCommandAnalyzer {
	return &TestCommandAnalyzer{}
}

func (a *TestCommandAnalyzer) AnalyzePrompt(prompt string, client APIClient) (*CommandProposal, bool) {
	// LLMにテストコマンドの推論を依頼
	messages := []ChatMessage{
		{
			Role: "system",
			Content: `あなたはテストコマンドを提案するエキスパートです。
ユーザーの入力からテストに関する意図を読み取り、適切なテストコマンドを提案してください。
テストコマンドを提案する場合は以下のフォーマットで返してください：

---COMMAND---
[実行するコマンド]
---EXPLANATION---
[コマンドの説明]
---END---

テストに関係ない場合は "NOT_TEST" と返してください。`,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := client.CreateChatCompletion(messages)
	if err != nil {
		return nil, false
	}

	if response == "NOT_TEST" {
		return nil, false
	}

	// レスポンスからコマンドと説明を抽出
	command, explanation, err := parseCommandResponse(response)
	if err != nil {
		return nil, false
	}

	return &CommandProposal{
		Command:     command,
		Explanation: explanation,
		Type:        "test",
	}, true
}

func parseCommandResponse(response string) (command string, explanation string, err error) {
	// デバッグ用にレスポンスを出力
	fmt.Printf("LLMからのレスポンス:\n%s\n", response)

	cmdStart := "---COMMAND---"
	expStart := "---EXPLANATION---"
	end := "---END---"

	// 改行を含む可能性を考慮
	response = strings.ReplaceAll(response, "\r\n", "\n")

	cmdIndex := strings.Index(response, cmdStart)
	expIndex := strings.Index(response, expStart)
	endIndex := strings.Index(response, end)

	if cmdIndex == -1 || expIndex == -1 || endIndex == -1 {
		return "", "", fmt.Errorf("invalid response format: missing markers")
	}

	cmdStartPos := cmdIndex + len(cmdStart)
	expStartPos := expIndex + len(expStart)

	// コマンドとその説明を抽出
	command = strings.TrimSpace(response[cmdStartPos:expIndex])
	explanation = strings.TrimSpace(response[expStartPos:endIndex])

	// 空のコマンドや説明をチェック
	if command == "" {
		return "", "", fmt.Errorf("empty command")
	}
	if explanation == "" {
		return "", "", fmt.Errorf("empty explanation")
	}

	return command, explanation, nil
}

// CommandExecutor はコマンドを実行するインターフェース
type CommandExecutor interface {
	Execute(command string) (string, error)
}

// DefaultCommandExecutor は実際のコマンドを実行する
type DefaultCommandExecutor struct{}

func NewDefaultCommandExecutor() *DefaultCommandExecutor {
	return &DefaultCommandExecutor{}
}

func (e *DefaultCommandExecutor) Execute(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	err := cmd.Run()
	return stdout.String() + stderr.String(), err
}

// CommandRunner はコマンドの実行を管理する
type CommandRunner struct {
	executor CommandExecutor
	client   APIClient
}

func NewCommandRunner(executor CommandExecutor, client APIClient) *CommandRunner {
	if executor == nil {
		executor = NewDefaultCommandExecutor()
	}
	return &CommandRunner{
		executor: executor,
		client:   client,
	}
}

// RunWithApproval はユーザーの承認を得てからコマンドを実行し、必要に応じて修正を提案する
func (r *CommandRunner) RunWithApproval(proposal *CommandProposal) error {
	if proposal == nil {
		return fmt.Errorf("コマンドの提案がありません")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("\n提案されたコマンド: %s\n", proposal.Command)
		fmt.Printf("説明: %s\n", proposal.Explanation)
		fmt.Print("このコマンドを実行しますか？ [y/N]: ")

		response, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// EOFの場合はデフォルトでNoとして扱う
				return fmt.Errorf("コマンドの実行がキャンセルされました")
			}
			return fmt.Errorf("入力の読み取りに失敗: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" {
			return fmt.Errorf("コマンドの実行がキャンセルされました")
		}

		fmt.Printf("\nコマンドを実行します: %s\n", proposal.Command)
		output, err := r.executor.Execute(proposal.Command)

		if err == nil {
			fmt.Println("\n✅ コマンドが正常に実行されました")
			return nil
		}

		// エラーが発生した場合、LLMに修正案を要求
		messages := []ChatMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`あなたはコマンドの実行エラーを解決するエキスパートです。
以下のコマンドがエラーで失敗しました。エラーを分析し、修正案を提示してください。

実行したコマンド：
%s

実行結果：
%s

応答は以下の2つのパターンのいずれかで返してください：

1. 修正案がある場合：
---COMMAND---
[修正したコマンド]
---EXPLANATION---
[修正内容の説明]
---END---

2. 修正の必要がない、またはコマンドでは解決できない場合：
NO_FIX_NEEDED

それ以外の形式は使用しないでください。`, proposal.Command, output),
			},
		}

		response, err = r.client.CreateChatCompletion(messages)
		if err != nil {
			return fmt.Errorf("APIリクエストに失敗: %w", err)
		}

		// デバッグ用にレスポンスを出力
		fmt.Printf("\nLLMからの応答:\n%s\n", response)

		if response == "NO_FIX_NEEDED" {
			return fmt.Errorf("コマンドの実行に失敗し、修正案はありません: %w", err)
		}

		// 新しいコマンドの提案を解析
		command, explanation, err := parseCommandResponse(response)
		if err != nil {
			return fmt.Errorf("修正案の解析に失敗: %w", err)
		}

		proposal = &CommandProposal{
			Command:     command,
			Explanation: explanation,
			Type:        proposal.Type,
		}

		fmt.Println("\n❌ コマンドの実行に失敗しました。修正案が提案されました。")
	}
}
