package makasero

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/samber/lo"
	"google.golang.org/api/option"
)

type Agent struct {
	client     *genai.Client
	model      *genai.GenerativeModel
	chat       *genai.ChatSession
	session    *Session
	functions  map[string]FunctionDefinition
	mcpManager *MCPClientManager
	debug      bool
	apiKey     string
	modelName  string
}

type AgentOption func(*Agent)

func WithDebug(debug bool) AgentOption {
	return func(a *Agent) {
		a.debug = debug
	}
}

func WithSession(session *Session) AgentOption {
	return func(a *Agent) {
		a.session = session
	}
}

func WithModelName(modelName string) AgentOption {
	return func(a *Agent) {
		a.modelName = modelName
	}
}

func NewAgent(ctx context.Context, apiKey string, config *MCPConfig, opts ...AgentOption) (*Agent, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	agent := &Agent{
		apiKey:    apiKey,
		modelName: "gemini-2.0-flash-lite", // default model
		functions: make(map[string]FunctionDefinition),
	}

	for _, opt := range opts {
		opt(agent)
	}

	mcpManager := NewMCPClientManager()
	if err := mcpManager.InitializeFromConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP clients: %v", err)
	}
	agent.mcpManager = mcpManager

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %v", err)
	}
	agent.client = client

	model := client.GenerativeModel(agent.modelName)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("You are an AI assistant.\n" +
				"Execute tasks from users and always call the 'complete' function when a task is finished.\n" +
				"When calling functions, do not write the function name as text, but actually call the function."),
		},
	}
	agent.model = model

	for name, fn := range builtinFunctions {
		agent.functions[name] = fn
	}

	mcpFuncDecls, err := mcpManager.GenerateAllFunctionDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate MCP tools: %v", err)
	}

	for _, fn := range mcpFuncDecls {
		agent.functions[fn.Declaration.Name] = fn
	}

	var allFuncDeclarations []*genai.FunctionDeclaration
	for _, fn := range agent.functions {
		allFuncDeclarations = append(allFuncDeclarations, fn.Declaration)
	}

	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: allFuncDeclarations,
		},
	}

	model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingAny,
		},
	}

	if agent.session == nil {
		agent.session = &Session{
			ID:        generateSessionID(),
			CreatedAt: time.Now(),
		}
	}

	agent.chat = model.StartChat()
	if len(agent.session.History) > 0 {
		agent.chat.History = agent.session.History
	}

	mcpManager.SetupNotificationHandlers(func(serverName string, notification mcp.JSONRPCNotification) {
		if agent.debug {
			fmt.Printf("[%s] Notification: %v\n", serverName, notification)
		} else {
			agent.handleNotification(notification)
		}
	})

	return agent, nil
}

func (a *Agent) Close() error {
	if a.client != nil {
		a.client.Close()
	}
	return nil
}

func (a *Agent) ProcessMessage(ctx context.Context, userInput string) error {
	fmt.Printf("\n--- Start session ---\n")
	fmt.Printf("\n🗣️ Sending message to AI:\n%s\n", strings.TrimSpace(userInput))

	resp, err := a.chat.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		a.session.History = a.chat.History
		a.session.UpdatedAt = time.Now()
		SaveSession(a.session)
		return fmt.Errorf("failed to send message to AI: %v", err)
	}

	for {
		if len(resp.Candidates) == 0 ||
			len(resp.Candidates[0].Content.Parts) == 0 {
			resp, err = a.chat.SendMessage(ctx, genai.Text("Task may not be finished. Please continue.\n"+
				"If you have finished the task, please call the 'complete' function.\n"+
				"If you have any questions, please call the 'askQuestion' function."))
			if err != nil {
				return fmt.Errorf("failed to send message to AI: %v", err)
			}
			fmt.Printf("\n🗣️ Please continue the conversation:\n")
		} else {
			break
		}
	}

	return a.processResponse(ctx, resp)
}

func (a *Agent) processResponse(ctx context.Context, resp *genai.GenerateContentResponse) error {
loop:
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			var functionCallingResponses []genai.FunctionResponse
			if a.debug {
				fmt.Printf("\n🔍 len parts: %d\n", len(cand.Content.Parts))
			}
			for _, part := range cand.Content.Parts {
				switch p := part.(type) {
				case genai.FunctionCall:
					fmt.Printf("\n🔧 AI uses function calling: %s\n", p.Name)

					if a.debug {
						buf, err := json.MarshalIndent(p, "", "  ")
						if err != nil {
							return fmt.Errorf("failed to marshal function response: %v", err)
						}
						fmt.Printf("\n🔍 Debug function call:\n%s\n", string(buf))
					}

					var result map[string]any
					if strings.HasPrefix(p.Name, "mcp_") {
						var err error
						result, err = a.mcpManager.CallMCPTool(ctx, p.Name, p.Args)
						if err != nil {
							return fmt.Errorf("MCP function %s failed: %v", p.Name, err)
						}
					} else {
						fn, exists := a.functions[p.Name]
						if !exists {
							return fmt.Errorf("unknown function: %s", p.Name)
						}

						var err error
						result, err = fn.Handler(ctx, p.Args)
						if err != nil {
							return fmt.Errorf("function %s failed: %v", p.Name, err)
						}

						if p.Name == "complete" || p.Name == "askQuestion" {
							fmt.Println("\n--- Finish session ---")
							a.session.History = a.chat.History
							a.session.UpdatedAt = time.Now()
							if err := SaveSession(a.session); err != nil {
								return err
							}
							fmt.Printf("Session ID: %s\n", a.session.ID)
							return nil
						}
					}

					if a.debug {
						buf, err := json.MarshalIndent(result, "", "  ")
						if err != nil {
							return fmt.Errorf("failed to marshal function response: %v", err)
						}
						fmt.Printf("\n🔍 Debug function result:\n%s\n", string(buf))
					}

					functionCallingResponses = append(functionCallingResponses, genai.FunctionResponse{
						Name:     p.Name,
						Response: result,
					})
				case genai.Text:
					fmt.Printf("\n🤖 Response from AI:\n%s\n", strings.TrimSpace(string(p)))
				default:
					fmt.Printf("unknown response type: %T\n", part)
				}
			}

			parts := lo.Map(functionCallingResponses, func(fnResp genai.FunctionResponse, _ int) genai.Part { return fnResp })
			var err error
			if a.debug {
				buf, err := json.MarshalIndent(parts, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal function response: %v", err)
				}
				fmt.Printf("\n🔍 Debug send message:\n%s\n", string(buf))
			}
			resp, err = a.chat.SendMessage(ctx, parts...)
			if err != nil {
				return fmt.Errorf("failed to send function response: %v", err)
			}

			if a.debug {
				buf, err := json.MarshalIndent(resp, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal function response: %v", err)
				}
				fmt.Printf("\n🔍 Debug received response:\n%s\n", string(buf))
			}

			goto loop
		} else {
			fmt.Printf("response content is nil\n")
		}
	}

	resp, err := a.chat.SendMessage(ctx, genai.Text("Task may not be finished. Please continue.\n"+
		"If you have finished the task, please call the 'complete' function.\n"+
		"If you have any questions, please call the 'askQuestion' function."))
	if err != nil {
		return fmt.Errorf("failed to send message to AI: %v", err)
	}
	fmt.Printf("\n🗣️ Please continue the conversation:\n")
	goto loop
}

func (a *Agent) GetSession() *Session {
	return a.session
}

func (a *Agent) LoadSession(sessionID string) error {
	session, err := LoadSession(sessionID)
	if err != nil {
		return err
	}
	a.session = session
	a.chat = a.model.StartChat()
	a.chat.History = session.History
	return nil
}

func (a *Agent) handleNotification(notification mcp.JSONRPCNotification) {
	fmt.Printf("Received notification: %v\n", notification)
}

func (a *Agent) GetAvailableFunctions() []string {
	var functionNames []string
	for name := range a.functions {
		functionNames = append(functionNames, name)
	}
	return functionNames
}

func (a *Agent) GetStderrReaders() map[string]io.Reader {
	return a.mcpManager.GetStderrReaders()
}
