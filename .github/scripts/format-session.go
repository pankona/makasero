package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type SerializablePart struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type SerializableContent struct {
	Parts []SerializablePart `json:"parts"`
	Role  string             `json:"role"`
}

type Session struct {
	ID                string                `json:"id"`
	SerializedHistory []SerializableContent `json:"history"`
}

func main() {
	data, err := os.ReadFile(".makasero/sessions/test_session.json")
	if err != nil {
		fmt.Printf("Error reading session file: %v\n", err)
		return
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		fmt.Printf("Error parsing session JSON: %v\n", err)
		return
	}

	fmt.Println("## 🤖 LLMチャット履歴")
	fmt.Println("")

	for i, content := range session.SerializedHistory {
		roleJP := content.Role
		if content.Role == "user" {
			roleJP = "ユーザー"
		} else if content.Role == "model" {
			roleJP = "AI"
		}
		fmt.Printf("### メッセージ %d (%s)\n", i+1, roleJP)

		for _, part := range content.Parts {
			switch part.Type {
			case "text":
				if text, ok := part.Content.(string); ok {
					fmt.Printf("```\n%s\n```\n", text)
				}
			case "function_call":
				if fc, ok := part.Content.(map[string]interface{}); ok {
					// Improved type safety - check if Name exists and is not nil
					if name, exists := fc["Name"]; exists && name != nil {
						if nameStr, ok := name.(string); ok {
							fmt.Printf("**関数呼び出し:** `%s`\n", nameStr)
						}
					}
					if args, exists := fc["Args"]; exists && args != nil {
						if argsMap, ok := args.(map[string]interface{}); ok && len(argsMap) > 0 {
							fmt.Printf("**引数:** `%v`\n", argsMap)
						}
					}
				}
			case "function_response":
				if fr, ok := part.Content.(map[string]interface{}); ok {
					// Improved type safety - check if Name exists and is not nil
					if name, exists := fr["Name"]; exists && name != nil {
						if nameStr, ok := name.(string); ok {
							fmt.Printf("**関数レスポンス:** `%s`\n", nameStr)
						}
					}
					if resp, exists := fr["Response"]; exists && resp != nil {
						if respMap, ok := resp.(map[string]interface{}); ok {
							if output, exists := respMap["output"]; exists && output != nil {
								if outputStr, ok := output.(string); ok && outputStr != "" {
									fmt.Printf("```\n%s\n```\n", outputStr)
								}
							}
						}
					}
				}
			}
		}
		fmt.Println("")
	}
}
