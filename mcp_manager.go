package makasero

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pankona/makasero/mlog"
)

type MCPClientManager struct {
	clients     map[string]*MCPClient
	clientsLock sync.RWMutex
}

func NewMCPClientManager() *MCPClientManager {
	return &MCPClientManager{
		clients: make(map[string]*MCPClient),
	}
}

func (m *MCPClientManager) InitializeFromConfig(ctx context.Context, config *MCPConfig) error {
	for serverName, serverConfig := range config.MCPServers {
		client, err := NewMCPClient(ServerCmd{
			Cmd:  serverConfig.Command,
			Args: serverConfig.Args,
			Env:  serverConfig.Env,
		})
		if err != nil {
			return fmt.Errorf("failed to create MCP client for %s: %v", serverName, err)
		}

		initResult, err := client.Initialize(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize MCP client for %s: %v", serverName, err)
		}

		mlog.Debugf(ctx, "%s mcp server initialize result: %s", serverName, initResult)

		m.clientsLock.Lock()
		m.clients[serverName] = client
		m.clientsLock.Unlock()
	}

	return nil
}

func (m *MCPClientManager) Close(ctx context.Context) error {
	var errs []string

	m.clientsLock.RLock()
	defer m.clientsLock.RUnlock()

	for name, client := range m.clients {
		if err := client.Close(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("failed to close MCP client %s: %v", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors occurred: %s", strings.Join(errs, "; "))
	}

	return nil
}

func (m *MCPClientManager) GetClient(name string) (*MCPClient, bool) {
	m.clientsLock.RLock()
	defer m.clientsLock.RUnlock()

	client, ok := m.clients[name]
	return client, ok
}

func (m *MCPClientManager) GetAllClients() map[string]*MCPClient {
	m.clientsLock.RLock()
	defer m.clientsLock.RUnlock()

	clients := make(map[string]*MCPClient, len(m.clients))
	for name, client := range m.clients {
		clients[name] = client
	}

	return clients
}

func (m *MCPClientManager) GenerateAllFunctionDefinitions(ctx context.Context) ([]FunctionDefinition, error) {
	var allFunctions []FunctionDefinition
	var errs []string

	m.clientsLock.RLock()
	defer m.clientsLock.RUnlock()

	for serverName, client := range m.clients {
		functions, err := client.GenerateFunctionDefinitions(ctx, serverName)
		if err != nil {
			errs = append(errs, fmt.Sprintf("failed to generate function definitions for %s: %v", serverName, err))
			continue
		}

		allFunctions = append(allFunctions, functions...)
	}

	if len(errs) > 0 {
		return allFunctions, fmt.Errorf("multiple errors occurred while generating function definitions: %s", strings.Join(errs, "; "))
	}

	return allFunctions, nil
}

func (m *MCPClientManager) SetupNotificationHandlers(handler func(serverName string, notification mcp.JSONRPCNotification)) {
	m.clientsLock.RLock()
	defer m.clientsLock.RUnlock()

	for serverName, client := range m.clients {
		serverNameCopy := serverName
		client.OnNotification(func(notification mcp.JSONRPCNotification) {
			handler(serverNameCopy, notification)
		})
	}
}

func (m *MCPClientManager) GetStderrReaders() map[string]io.Reader {
	m.clientsLock.RLock()
	defer m.clientsLock.RUnlock()

	readers := make(map[string]io.Reader, len(m.clients))
	for name, client := range m.clients {
		readers[name] = client.Stderr()
	}

	return readers
}

func (m *MCPClientManager) CallMCPTool(ctx context.Context, fullName string, args map[string]any) (map[string]any, error) {
	parts := strings.SplitN(fullName, "_", 3)
	if len(parts) < 3 || parts[0] != "mcp" {
		return nil, fmt.Errorf("invalid MCP tool name format: %s", fullName)
	}

	serverName := parts[1]
	toolName := parts[2]

	m.clientsLock.RLock()
	client, ok := m.clients[serverName]
	m.clientsLock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("MCP server not found: %s", serverName)
	}

	result, err := client.callMCPTool(toolName, args)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]any); ok {
		return resultMap, nil
	}

	if mcpResult, ok := result.(*mcp.CallToolResult); ok {
		var contents []string
		for _, content := range mcpResult.Content {
			if textContent, ok := content.(mcp.TextContent); ok {
				contents = append(contents, textContent.Text)
			} else {
				contents = append(contents, fmt.Sprintf("%v", content))
			}
		}

		resultMap := map[string]any{
			"is_error": mcpResult.IsError,
			"content":  strings.Join(contents, "\n"),
		}
		if mcpResult.Result.Meta != nil {
			resultMap["meta"] = mcpResult.Result.Meta
		}

		return resultMap, nil
	}

	return nil, fmt.Errorf("unexpected result type from callMCPTool: %T", result)
}

func (m *MCPClientManager) GetFunctionDeclarations() ([]*genai.FunctionDeclaration, error) {
	functions, err := m.GenerateAllFunctionDefinitions(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to generate function definitions: %v", err)
	}

	declarations := make([]*genai.FunctionDeclaration, 0, len(functions))
	for _, fn := range functions {
		declarations = append(declarations, fn.Declaration)
	}

	return declarations, nil
}
