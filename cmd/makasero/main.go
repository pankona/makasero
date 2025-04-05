package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pankona/makasero/agent"
	"github.com/pankona/makasero/tools"
	"github.com/spf13/cobra"
)

func main() {
	var apiKey string

	rootCmd := &cobra.Command{
		Use:   "makasero",
		Short: "AI agent CLI tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Agentの作成
			agent, err := agent.New(apiKey)
			if err != nil {
				return fmt.Errorf("failed to create agent: %v", err)
			}
			defer agent.Close()

			// ツールの登録
			agent.RegisterTool(&tools.ExecCommand{})

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

			return nil
		},
	}

	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "Gemini API key")
	rootCmd.MarkFlagRequired("api-key")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
