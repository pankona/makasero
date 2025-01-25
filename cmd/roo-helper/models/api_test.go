package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRequestSerialization(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		want    map[string]interface{}
	}{
		{
			name: "basic request",
			request: Request{
				Command: "test",
				Input:   "input data",
			},
			want: map[string]interface{}{
				"command": "test",
				"input":   "input data",
			},
		},
		{
			name: "request with options",
			request: Request{
				Command: "test",
				Input:   "input data",
				Options: map[string]interface{}{"key": "value"},
			},
			want: map[string]interface{}{
				"command": "test",
				"input":   "input data",
				"options": map[string]interface{}{"key": "value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("Failed to marshal Request: %v", err)
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("Request serialization mismatch.\nGot: %v\nWant: %v", gotMap, tt.want)
			}
		})
	}
}

func TestResponseSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		want     map[string]interface{}
	}{
		{
			name: "success response",
			response: Response{
				Success: true,
				Data:    "test data",
			},
			want: map[string]interface{}{
				"success": true,
				"data":    "test data",
			},
		},
		{
			name: "error response",
			response: Response{
				Success: false,
				Error:   "test error",
			},
			want: map[string]interface{}{
				"success": false,
				"error":   "test error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.response)
			if err != nil {
				t.Errorf("Failed to marshal Response: %v", err)
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("Response serialization mismatch.\nGot: %v\nWant: %v", gotMap, tt.want)
			}
		})
	}
}

func TestChatMessageSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message ChatMessage
		want    map[string]interface{}
	}{
		{
			name: "user message",
			message: ChatMessage{
				Role:    "user",
				Content: "Hello",
			},
			want: map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		},
		{
			name: "system message",
			message: ChatMessage{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
			want: map[string]interface{}{
				"role":    "system",
				"content": "You are a helpful assistant",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.message)
			if err != nil {
				t.Errorf("Failed to marshal ChatMessage: %v", err)
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("ChatMessage serialization mismatch.\nGot: %v\nWant: %v", gotMap, tt.want)
			}
		})
	}
}

func TestChatRequestSerialization(t *testing.T) {
	tests := []struct {
		name    string
		request ChatRequest
		want    map[string]interface{}
	}{
		{
			name: "basic chat request",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			want: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("Failed to marshal ChatRequest: %v", err)
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("ChatRequest serialization mismatch.\nGot: %v\nWant: %v", gotMap, tt.want)
			}
		})
	}
}
