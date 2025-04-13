package main

import (
	"context"

	"github.com/google/generative-ai-go/genai"
)

type FunctionHandler func(ctx context.Context, args map[string]any) (map[string]any, error)

type FunctionDefinition struct {
	Declaration *genai.FunctionDeclaration
	Handler     FunctionHandler
}

var functions = map[string]FunctionDefinition{
	"complete": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "complete",
			Description: "タスク完了を報告します",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"message": {
						Type:        genai.TypeString,
						Description: "完了メッセージ",
					},
				},
				Required: []string{"message"},
			},
		},
		Handler: handleComplete,
	},
	"askQuestion": {
		Declaration: &genai.FunctionDeclaration{
			Name:        "askQuestion",
			Description: "ユーザーに質問を投げかけます",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"question": {
						Type:        genai.TypeString,
						Description: "ユーザーへの質問内容",
					},
					"options": {
						Type:        genai.TypeArray,
						Description: "選択肢（オプション）",
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
				Required: []string{"question"},
			},
		},
		Handler: handleAskQuestion,
	},
}

func handleComplete(ctx context.Context, args map[string]any) (map[string]any, error) {
	// 呼び出されないので実装しなくてよい
	return nil, nil
}

func handleAskQuestion(ctx context.Context, args map[string]any) (map[string]any, error) {
	// 呼び出されないので実装しなくてよい
	return nil, nil
}
