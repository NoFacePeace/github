package anthropic

import (
	"context"
	"encoding/json"

	ai "github.com/anthropics/anthropic-sdk-go"
)

// Tool 是调用方注册的工具。
type Tool struct {
	Name        string
	Description string
	// InputSchema 是工具输入参数的 JSON Schema properties，形如
	// map[string]any{"location": map[string]any{"type": "string", "description": "..."}}。
	InputSchema map[string]any
	// Required 是必填参数名。
	Required []string
	// Call 处理一次工具调用，input 是模型生成的输入参数（原始 JSON）。
	Call func(ctx context.Context, input json.RawMessage) (string, error)
}

// ToolRegistry 按名称索引工具。
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 构造注册表。
func NewToolRegistry(tools ...Tool) *ToolRegistry {
	r := &ToolRegistry{tools: make(map[string]Tool, len(tools))}
	for _, t := range tools {
		r.tools[t.Name] = t
	}
	return r
}

// Get 按名称获取工具。
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// toToolParams 将注册的工具转换为官方 SDK 的 ToolUnionParam 列表。
func (r *ToolRegistry) toToolParams() []ai.ToolUnionParam {
	if r == nil || len(r.tools) == 0 {
		return nil
	}
	out := make([]ai.ToolUnionParam, 0, len(r.tools))
	for _, t := range r.tools {
		tp := ai.ToolParam{
			Name:        t.Name,
			Description: ai.String(t.Description),
			InputSchema: ai.ToolInputSchemaParam{
				Properties: t.InputSchema,
				Required:   t.Required,
			},
		}
		out = append(out, ai.ToolUnionParam{OfTool: &tp})
	}
	return out
}
