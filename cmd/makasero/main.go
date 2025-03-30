package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pankona/makasero/agent"
)

var (
	apiKey = flag.String("api-key", "", "Gemini API Key")
)

func main() {
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("api-key is required")
	}

	// コンテキストの作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Geminiクライアントの作成
	client, err := agent.NewGeminiClient(ctx, *apiKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	// エージェントの作成
	a, err := agent.NewAgent(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	defer a.Close()

	fmt.Println("Makasero - AI Code Analysis Agent")
	fmt.Println("Type 'exit' to quit")
	fmt.Println()

	// 対話ループ
	for {
		fmt.Print("> ")
		var input string
		fmt.Scanln(&input)

		if input == "exit" {
			break
		}

		// エージェントにメッセージを送信
		response, err := a.SendMessage(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(response)
		fmt.Println()
	}
}
