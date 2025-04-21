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
	"github.com/pankona/makasero/mlog"
	"github.com/samber/lo"
	"google.golang.org/api/option"
)

func mustMarshalIndent(v interface{}) []byte {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return buf
}

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

	if agent.debug {
		mlog.ConfigureDebug()
	} else {
		mlog.Configure(mlog.DefaultConfig())
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
			Mode: genai.FunctionCallingAuto,
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
			mlog.Debugf(ctx, "[%s] Notification: %v", serverName, notification)
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
	ctx = mlog.ContextWithSessionID(ctx, a.session.ID)

	if a.debug {
		ctx = mlog.ContextWithDebug(ctx)
	}

	mlog.Infof(ctx, "\n--- Start session ---\n")
	mlog.Infof(ctx, "\n🗣️ Sending message to AI:\n%s\n", strings.TrimSpace(userInput))

	resp, err := a.chat.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		return fmt.Errorf("failed to send message to AI: %v", err)
	}

	mlog.Debugf(ctx, "\n🔍 Debug received response:\n%s\n", string(mustMarshalIndent(resp)))

	// continue loop until shouldStop is true
	for {
		newResp, shouldStop, err := a.processResponse(ctx, resp)
		if err != nil {
			return fmt.Errorf("failed to process response: %v", err)
		}

		if shouldStop {
			break
		}

		// shouldStop が false で resp が nil ということはまだタスクが終わっていないので続けてもらう
		if newResp == nil {
			newResp, err := a.chat.SendMessage(ctx,
				genai.Text("Task may not be finished. Please continue.\n"+
					"If you have finished the task, please call the 'complete' function.\n"+
					"If you have any questions, please call the 'askQuestion' function."))
			if err != nil {
				mlog.Errorf(ctx, "Failed to send message to AI: %v", err)
				return fmt.Errorf("failed to send message to AI: %v", err)
			}

			mlog.Infof(ctx, "\n🗣️ Please continue the conversation:\n")
			resp = newResp

			continue
		}

		resp = newResp
	}

	mlog.Infof(ctx, "\n--- Finish session ---")
	a.session.History = a.chat.History
	a.session.UpdatedAt = time.Now()
	if err := SaveSession(a.session); err != nil {
		mlog.Errorf(ctx, "Failed to save session: %v", err)
		return fmt.Errorf("failed to save session: %v", err)
	}
	mlog.Infof(ctx, "Session ID: %s", a.session.ID)

	return nil
}

func (a *Agent) processResponse(ctx context.Context, resp *genai.GenerateContentResponse) (*genai.GenerateContentResponse, bool, error) {
	ctx = mlog.ContextWithSessionID(ctx, a.session.ID)

	if a.debug {
		ctx = mlog.ContextWithDebug(ctx)
	}

	var functionCallingResponses []genai.FunctionResponse

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			mlog.Debugf(ctx, "\n🔍 len parts: %d\n", len(cand.Content.Parts))
			for _, part := range cand.Content.Parts {
				switch p := part.(type) {
				case genai.FunctionCall:
					fnCtx := mlog.ContextWithAttr(ctx, "function", p.Name)
					mlog.Infof(fnCtx, "\n🔧 AI uses function calling: %s\n", p.Name)

					mlog.Debugf(fnCtx, "\n🔍 Debug function call:\n%s\n", string(mustMarshalIndent(p)))

					var result map[string]any
					if strings.HasPrefix(p.Name, "mcp_") {
						// mcp functions
						var err error
						result, err = a.mcpManager.CallMCPTool(ctx, p.Name, p.Args)
						if err != nil {
							mlog.Errorf(ctx, "MCP function %s failed: %v", p.Name, err)
							result = map[string]any{
								"is_error": true,
								"output":   fmt.Sprintf("MCP function %s failed: %v", p.Name, err),
							}
						}
					} else {
						// builtin functions
						fn, exists := a.functions[p.Name]
						if !exists {
							mlog.Errorf(ctx, "Unknown function: %s", p.Name)
							result = map[string]any{
								"is_error": true,
								"output":   fmt.Sprintf("unknown function: %s", p.Name),
							}
						}

						var err error
						result, err = fn.Handler(ctx, p.Args)
						if err != nil {
							mlog.Errorf(ctx, "Function %s failed: %v", p.Name, err)
							result = map[string]any{
								"is_error": true,
								"output":   fmt.Sprintf("function %s failed: %v", p.Name, err),
							}
						}

						if p.Name == "complete" || p.Name == "askQuestion" {
							return nil, true, nil
						}
					}

					mlog.Debugf(ctx, "\n🔍 Debug function result:\n%s\n", string(mustMarshalIndent(result)))
					functionCallingResponses = append(functionCallingResponses, genai.FunctionResponse{
						Name:     p.Name,
						Response: result,
					})
				case genai.Text:
					mlog.Infof(ctx, "\n🤖 Response from AI:\n%s\n", strings.TrimSpace(string(p)))
				default:
					mlog.Warnf(ctx, "Unknown response type: %T", part)
				}
			}

			if len(functionCallingResponses) > 0 {
				parts := lo.Map(functionCallingResponses, func(fnResp genai.FunctionResponse, _ int) genai.Part { return fnResp })

				var err error
				mlog.Debugf(ctx, "\n🔍 Debug send message:\n%s\n", string(mustMarshalIndent(parts)))
				resp, err = a.chat.SendMessage(ctx, parts...)
				if err != nil {
					mlog.Errorf(ctx, "Failed to send function response: %v", err)
					return nil, false, fmt.Errorf("failed to send function response: %v", err)
				}

				mlog.Debugf(ctx, "\n🔍 Debug received response:\n%s\n", string(mustMarshalIndent(resp)))
				return resp, false, nil
			}
		} else {
			mlog.Warnf(ctx, "Response content is nil")
		}
	}

	return nil, true, nil
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
	ctx := context.Background()
	ctx = mlog.ContextWithSessionID(ctx, a.session.ID)
	mlog.Infof(ctx, "Received notification: %v", notification)
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
