package openai

import "encoding/json"

// 以下结构仅用于转换 Chat Completions 协议，不暴露给 Loop。
type completionRequest struct {
	Model    string         `json:"model"`
	Messages []messageParam `json:"messages"`
	Tools    []toolParam    `json:"tools,omitempty"`
	Stream   bool           `json:"stream"`
}

type messageParam struct {
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	ToolCalls  []callParam `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type callParam struct {
	Index    int          `json:"index,omitempty"`
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function functionCall `json:"function"`
}

type functionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments"`
}

type toolParam struct {
	Type     string             `json:"type"`
	Function functionDefinition `json:"function"`
}

type functionDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type completionChunk struct {
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content   string      `json:"content"`
			Refusal   string      `json:"refusal"`
			ToolCalls []callParam `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}
