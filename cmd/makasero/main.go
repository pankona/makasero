package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pankona/makasero/makasero"
)

var (
	debug          = flag.Bool("debug", false, "debug mode")
	promptFile     = flag.String("f", "", "prompt file")
	configFilePath = flag.String("config", "", "path to config file")
	listSessionsFlag = flag.Bool("ls", false, "利用可能なセッション一覧を表示")
	sessionID        = flag.String("s", "", "継続するセッションID")
	showHistory      = flag.String("sh", "", "指定したセッションIDの会話履歴全文を表示")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func readPromptFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %v", err)
	}
	return string(content), nil
}

func run() error {
	// コマンドライン引数の処理
	flag.Parse()

	// セッション一覧表示の処理
	if *listSessionsFlag {
		return makasero.PrintSessionsList()
	}

	// 会話履歴全文表示の処理
	if *showHistory != "" {
		return makasero.PrintSessionHistory(*showHistory)
	}

	// 設定ファイルの読み込み
	config, err := makasero.LoadConfig(*configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load config: %v\nPlease create a config file at ~/.makasero/config.json with your MCP server settings", err)
	}

	// プロンプトの取得
	args := flag.Args()
	var userInput string
	if *promptFile != "" {
		// ファイルからプロンプトを読み込む
		prompt, err := readPromptFromFile(*promptFile)
		if err != nil {
			return err
		}
		userInput = prompt
	} else if len(args) > 0 {
		// コマンドライン引数からプロンプトを取得
		userInput = strings.Join(args, " ")
	} else {
		return fmt.Errorf("Please specify a prompt (command line arguments or -f option)")
	}

	// APIキーの取得
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	// モデル名の取得（デフォルト: gemini-2.0-flash-lite）
	modelName := os.Getenv("MODEL_NAME")

	// コンテキストの作成
	ctx := context.Background()

	// エージェントの作成
	agentOptions := []makasero.AgentOption{
		makasero.WithDebug(*debug),
	}

	// 既存のセッションを読み込む場合
	if *sessionID != "" {
		session, err := makasero.LoadSession(*sessionID)
		if err != nil {
			return err
		}
		agentOptions = append(agentOptions, makasero.WithSession(session))
	}

	// モデル名が指定されている場合
	if modelName != "" {
		agentOptions = append(agentOptions, makasero.WithModelName(modelName))
	}

	// エージェントの初期化
	agent, err := makasero.NewAgent(ctx, apiKey, config, agentOptions...)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %v", err)
	}
	defer agent.Close()

	// 標準エラー出力のキャプチャ
	stderrReaders := agent.GetStderrReaders()
	for serverName, reader := range stderrReaders {
		serverNameCopy := serverName
		go func(r io.Reader) {
			if *debug {
				buf := make([]byte, 1024)
				for {
					n, err := r.Read(buf)
					if err != nil {
						if err != io.EOF {
							fmt.Fprintf(os.Stderr, "[%s] stderr read error: %v\n", serverNameCopy, err)
						}
						return
					}
					fmt.Fprintf(os.Stderr, "[%s] %s", serverNameCopy, buf[:n])
				}
			} else {
				io.Copy(os.Stderr, r)
			}
		}(reader)
	}

	// 利用可能な関数の一覧表示
	fmt.Printf("declared tools: %d\n", len(agent.GetAvailableFunctions()))
	for _, name := range agent.GetAvailableFunctions() {
		fmt.Printf("%s\n", name)
	}

	// メッセージの処理
	if err := agent.ProcessMessage(ctx, userInput); err != nil {
		return err
	}

	return nil
}
