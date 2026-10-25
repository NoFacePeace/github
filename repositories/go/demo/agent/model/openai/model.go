package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

type Config struct {
	APIKey     string
	Model      string
	BaseURL    string // 默认为 https://api.openai.com/v1
	HTTPClient *http.Client
}

type Model struct {
	config Config
}

var _ model.Model = (*Model)(nil)

func New(config Config) *Model {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	return &Model{config: config}
}

func (m *Model) Generate(ctx context.Context, req model.Request, onEvent func(model.Event)) error {
	payload := completionRequest{Model: m.config.Model, Stream: true}
	for _, message := range req.Messages {
		converted := messageParam{
			Role: string(message.Role), Content: message.Content, ToolCallID: message.ToolCallID,
		}
		for _, call := range message.ToolCalls {
			converted.ToolCalls = append(converted.ToolCalls, callParam{
				ID: call.ID, Type: "function",
				Function: functionCall{Name: call.Name, Arguments: string(call.Arguments)},
			})
		}
		payload.Messages = append(payload.Messages, converted)
	}
	for _, definition := range req.Tools {
		payload.Tools = append(payload.Tools, toolParam{
			Type: "function", Function: functionDefinition{
				Name: definition.Name, Description: definition.Description, Parameters: definition.InputSchema,
			},
		})
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode completion request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(m.config.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create completion request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+m.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	response, err := m.config.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request completion: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
		if err != nil {
			return fmt.Errorf("read completion error: %w", err)
		}
		return fmt.Errorf("completion HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return readStream(response.Body, onEvent)
}

func readStream(reader io.Reader, onEvent func(model.Event)) error {
	var text strings.Builder
	calls := make(map[int]*callParam)
	finishReason := ""
	// SSE 以空行分隔事件，多行 data 字段按换行符连接。
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	var data []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
			continue
		}
		if line != "" || len(data) == 0 {
			continue
		}
		payload := strings.Join(data, "\n")
		data = nil
		if payload == "[DONE]" {
			if finishReason != "stop" && finishReason != "tool_calls" {
				return fmt.Errorf("completion stopped unexpectedly: %q", finishReason)
			}
			message := model.Message{Role: model.RoleAssistant, Content: text.String()}
			indices := make([]int, 0, len(calls))
			for index := range calls {
				indices = append(indices, index)
			}
			sort.Ints(indices)
			for _, index := range indices {
				call := calls[index]
				message.ToolCalls = append(message.ToolCalls, tool.Call{
					ID: call.ID, Name: call.Function.Name, Arguments: json.RawMessage(call.Function.Arguments),
				})
			}
			onEvent(model.CompletedEvent{Message: message})
			return nil
		}
		var chunk completionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return fmt.Errorf("decode completion chunk: %w", err)
		}
		if chunk.Error != nil {
			return fmt.Errorf("completion error: %s", chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			if choice.Index != 0 {
				continue
			}
			if choice.Delta.Content != "" {
				text.WriteString(choice.Delta.Content)
				onEvent(model.TextDeltaEvent{Text: choice.Delta.Content})
			}
			if choice.Delta.Refusal != "" {
				text.WriteString(choice.Delta.Refusal)
				onEvent(model.TextDeltaEvent{Text: choice.Delta.Refusal})
			}
			for _, delta := range choice.Delta.ToolCalls {
				if calls[delta.Index] == nil {
					calls[delta.Index] = &callParam{}
				}
				call := calls[delta.Index]
				call.ID += delta.ID
				call.Function.Name += delta.Function.Name
				call.Function.Arguments += delta.Function.Arguments
			}
			if choice.FinishReason != "" {
				finishReason = choice.FinishReason
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read completion stream: %w", err)
	}
	return io.ErrUnexpectedEOF
}
