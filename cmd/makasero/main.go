package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pankona/makasero/agent"
	"github.com/pankona/makasero/tools"
)

func main() {
	// APIキーの設定
	apiKey := os.Getenv("GEMINI_API_KEY")

	// Agentの作成
	agent, err := agent.New(apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %v\n", err)
		os.Exit(1)
	}
	defer agent.Close()

	// ツールの登録
	agent.RegisterTool(&tools.ExecCommand{})
	agent.RegisterTool(&tools.CompleteTool{})

	// 対話モードの開始
	fmt.Println("makasero started. Type 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			break
		}

		// 入力の処理
		response, err := agent.Process(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(response)
	}
}
