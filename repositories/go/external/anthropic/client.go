package anthropic

import (
	"context"
	"fmt"

	ai "github.com/anthropics/anthropic-sdk-go"
	aiopt "github.com/anthropics/anthropic-sdk-go/option"
)

// Client 在官方 anthropic SDK 之上提供高层封装。
type Client struct {
	sdk       ai.Client
	model     string
	maxTokens int64
}

// New 构造 Client。API Key 默认取环境变量 ANTHROPIC_API_KEY。
func New(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o.apply(&cfg)
	}

	if cfg.apiKey == "" {
		return nil, fmt.Errorf("anthropic client init error: [api key is empty, set ANTHROPIC_API_KEY or use WithAPIKey]")
	}

	sdkOpts := []aiopt.RequestOption{
		aiopt.WithAPIKey(cfg.apiKey),
	}
	if cfg.baseURL != "" {
		sdkOpts = append(sdkOpts, aiopt.WithBaseURL(cfg.baseURL))
	}
	if cfg.timeout > 0 {
		sdkOpts = append(sdkOpts, aiopt.WithRequestTimeout(cfg.timeout))
	}
	if cfg.maxRetries > 0 {
		sdkOpts = append(sdkOpts, aiopt.WithMaxRetries(cfg.maxRetries))
	}

	return &Client{
		sdk:       ai.NewClient(sdkOpts...),
		model:     cfg.model,
		maxTokens: cfg.maxTokens,
	}, nil
}

// CompleteParams 是非流式对话的输入参数。
type CompleteParams struct {
	Model       string
	MaxTokens   int64
	System      string
	Messages    []Message
	Tools       []Tool
	Temperature float64
	StopSeqs    []string
}

// CompleteResponse 是非流式对话的结果。
type CompleteResponse struct {
	// Message 是官方返回的原始消息，可用于读取 StopReason、Usage 等。
	Message *ai.Message
	Text    string
	Blocks  []Block
}

// Complete 执行一次非流式 Messages.New 调用。
func (c *Client) Complete(ctx context.Context, params CompleteParams) (*CompleteResponse, error) {
	sdkMessages, err := toSDKMessages(params.Messages)
	if err != nil {
		return nil, err
	}
	sdkParams, err := c.buildParams(params.Model, params.MaxTokens, params.System, sdkMessages, params.Tools, params.Temperature, params.StopSeqs)
	if err != nil {
		return nil, err
	}
	msg, err := c.sdk.Messages.New(ctx, sdkParams)
	if err != nil {
		return nil, fmt.Errorf("complete messages error: [%w]", err)
	}
	blocks, err := extractBlocks(msg)
	if err != nil {
		return nil, err
	}
	return &CompleteResponse{
		Message: msg,
		Text:    extractText(msg),
		Blocks:  blocks,
	}, nil
}

// buildParams 把高层参数组装为官方 MessageNewParams，供 Complete/Stream/AgentLoop 共用。
// sdkMessages 是已转换好的官方 MessageParam 列表。
func (c *Client) buildParams(
	model string,
	maxTokens int64,
	system string,
	sdkMessages []ai.MessageParam,
	tools []Tool,
	temperature float64,
	stopSeqs []string,
) (ai.MessageNewParams, error) {
	model = strOrDefault(model, c.model)
	if model == "" {
		return ai.MessageNewParams{}, fmt.Errorf("build params error: [model is empty]")
	}
	if maxTokens <= 0 {
		maxTokens = c.maxTokens
	}

	p := ai.MessageNewParams{
		Model:     ai.Model(model),
		MaxTokens: maxTokens,
		Messages:  sdkMessages,
	}
	if system != "" {
		p.System = []ai.TextBlockParam{{Text: system}}
	}
	if len(stopSeqs) > 0 {
		p.StopSequences = stopSeqs
	}
	if temperature > 0 {
		p.Temperature = ai.Float(temperature)
	}
	if len(tools) > 0 {
		p.Tools = NewToolRegistry(tools...).toToolParams()
	}
	return p, nil
}

func strOrDefault(s, def string) string {
	if s != "" {
		return s
	}
	return def
}
