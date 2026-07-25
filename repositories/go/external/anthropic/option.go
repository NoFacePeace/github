package anthropic

import (
	"os"
	"time"
)

// Option 配置 Client 的构造选项。
type Option interface {
	apply(*config)
}

type config struct {
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int64
	timeout    time.Duration
	maxRetries int
}

type optionFunc func(*config)

func (f optionFunc) apply(c *config) { f(c) }

// WithAPIKey 设置 API Key。未设置时默认取环境变量 ANTHROPIC_API_KEY。
func WithAPIKey(key string) Option {
	return optionFunc(func(c *config) { c.apiKey = key })
}

// WithBaseURL 覆盖 API 基地址（用于代理或兼容网关）。
func WithBaseURL(url string) Option {
	return optionFunc(func(c *config) { c.baseURL = url })
}

// WithDefaultModel 设置默认模型，Complete/Stream/AgentLoop 未指定模型时使用。
func WithDefaultModel(model string) Option {
	return optionFunc(func(c *config) { c.model = model })
}

// WithDefaultMaxTokens 设置默认 max_tokens。SDK 的 MaxTokens 字段是 int64。
func WithDefaultMaxTokens(tokens int64) Option {
	return optionFunc(func(c *config) { c.maxTokens = tokens })
}

// WithTimeout 设置单次请求超时。
func WithTimeout(d time.Duration) Option {
	return optionFunc(func(c *config) { c.timeout = d })
}

// WithMaxRetries 设置最大重试次数。
func WithMaxRetries(n int) Option {
	return optionFunc(func(c *config) { c.maxRetries = n })
}

func defaultConfig() config {
	return config{
		apiKey:    os.Getenv("ANTHROPIC_API_KEY"),
		model:     defaultModel,
		maxTokens: defaultMaxTokens,
	}
}

const (
	defaultModel     = "claude-sonnet-5-5"
	defaultMaxTokens = 1024
)
