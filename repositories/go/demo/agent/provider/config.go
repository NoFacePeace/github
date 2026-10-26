package provider

import "net/http"

// Config 保存供应商的连接和认证配置，可由不同协议客户端共用。
type Config struct {
	BaseURL    string
	APIKey     string
	Headers    http.Header // 额外请求头，可覆盖协议客户端的默认请求头
	HTTPClient *http.Client
}
