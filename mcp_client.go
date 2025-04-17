package makasero

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type ServerCmd struct {
	Cmd  string
	Args []string
	Env  map[string]string
}

type MCPClient struct {
	client *client.StdioMCPClient
}

func NewMCPClient(serverCmd ServerCmd) (*MCPClient, error) {
	var env []string
	if serverCmd.Env != nil {
		env = expandEnvVars(serverCmd.Env)
	}
	
	client, err := client.NewStdioMCPClient(
		serverCmd.Cmd,
		env,
		serverCmd.Args...,
	)
	if err != nil {
		return nil, err
	}
	return &MCPClient{client: client}, nil
}

func (c *MCPClient) Close(ctx context.Context) error {
	return c.client.Close()
}

func (c *MCPClient) Stderr() io.Reader {
	return c.client.Stderr()
}

type InitializeResult string

func (c *MCPClient) Initialize(ctx context.Context) (InitializeResult, error) {
	result, err := c.client.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		return "", err
	}

	ret := mustMarshalIndent(result)
	return InitializeResult(ret), nil
}

func (c *MCPClient) GenerateFunctionDefinitions(ctx context.Context, prefix string) ([]FunctionDefinition, error) {
	tools, err := c.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	ret := make([]FunctionDefinition, 0, len(tools.Tools))
	for _, tool := range tools.Tools {
		ret = append(ret, FunctionDefinition{
			Declaration: &genai.FunctionDeclaration{
				Name:        fmt.Sprintf("mcp_%s_%s", prefix, tool.Name),
				Description: tool.Description,
				Parameters: &genai.Schema{
					Type:       genai.TypeObject,
					Properties: c.convertMCPParameters(tool.InputSchema),
				},
			},
			Handler: func(ctx context.Context, args map[string]any) (map[string]any, error) {
				result, err := c.callMCPTool(tool.Name, args)
				if err != nil {
					return nil, err
				}

				mcpResult, ok := result.(*mcp.CallToolResult)
				if !ok {
					return nil, fmt.Errorf("unexpected result type: %T", result)
				}

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
			},
		})
	}

	return ret, nil
}

func (c *MCPClient) OnNotification(handler func(notification mcp.JSONRPCNotification)) {
	c.client.OnNotification(handler)
}

func (c *MCPClient) callMCPTool(name string, args map[string]any) (interface{}, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	result, err := c.client.CallTool(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *MCPClient) convertMCPParameters(schema mcp.ToolInputSchema) map[string]*genai.Schema {
	converted := make(map[string]*genai.Schema)
	if schema.Properties == nil {
		return converted
	}

	for name, p := range schema.Properties {
		prop, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		typeVal, ok := prop["type"].(string)
		if !ok {
			continue
		}

		description := ""
		if desc, ok := prop["description"].(string); ok {
			description = desc
		}

		schema := &genai.Schema{
			Type:        c.convertSchemaType(typeVal),
			Description: description,
		}

		if typeVal == "array" {
			if items, ok := prop["items"].(map[string]interface{}); ok {
				itemType, hasType := items["type"].(string)
				itemDesc, hasDesc := items["description"].(string)
				if hasType {
					schema.Items = &genai.Schema{
						Type: c.convertSchemaType(itemType),
					}
					if hasDesc {
						schema.Items.Description = itemDesc
					}
				}
			}
		}

		if typeVal == "object" {
			if properties, ok := prop["properties"].(map[string]interface{}); ok {
				schema.Properties = make(map[string]*genai.Schema)
				for subName, subProp := range properties {
					if subPropMap, ok := subProp.(map[string]interface{}); ok {
						subType, hasType := subPropMap["type"].(string)
						subDesc, hasDesc := subPropMap["description"].(string)
						if hasType {
							schema.Properties[subName] = &genai.Schema{
								Type: c.convertSchemaType(subType),
							}
							if hasDesc {
								schema.Properties[subName].Description = subDesc
							}
						}
					}
				}
			}
			if required, ok := prop["required"].([]interface{}); ok {
				schema.Required = make([]string, 0, len(required))
				for _, r := range required {
					if str, ok := r.(string); ok {
						schema.Required = append(schema.Required, str)
					}
				}
			}
		}

		converted[name] = schema
	}

	return converted
}

func (c *MCPClient) convertSchemaType(schemaType string) genai.Type {
	switch schemaType {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeString // デフォルトは string
	}
}

func expandEnvVars(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for key, value := range env {
		expandedValue := os.ExpandEnv(value)
		result = append(result, key+"="+expandedValue)
	}
	return result
}
