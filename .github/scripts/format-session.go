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
	// Read session file
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

	// Print header in Japanese
	fmt.Println("## 🤖 LLMチャット履歴")
	fmt.Println("")

	for i, content := range session.SerializedHistory {
		fmt.Printf("### メッセージ %d (%s)\n", i+1, content.Role)

		for _, part := range content.Parts {
			switch part.Type {
			case "text":
				if textContent, ok := part.Content.(string); ok {
					fmt.Printf("```\n%s\n```\n", textContent)
				}
			case "function_call":
				if fc, ok := part.Content.(map[string]interface{}); ok {
					// Improved type safety for function calls
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
					// Improved type safety for function responses
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