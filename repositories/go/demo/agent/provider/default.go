package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultProvider 使用 OpenAI 兼容的 Models API 获取模型列表。
type DefaultProvider struct {
	config Config
}

var _ Provider = (*DefaultProvider)(nil)

func New(config Config) *DefaultProvider {
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	return &DefaultProvider{config: config}
}

func (p *DefaultProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimRight(p.config.BaseURL, "/")+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("create models request: %w", err)
	}
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}
	req.Header.Set("Accept", "application/json")
	for name, values := range p.config.Headers {
		req.Header.Del(name)
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	response, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
		if err != nil {
			return nil, fmt.Errorf("read models error: %w", err)
		}
		return nil, fmt.Errorf("models HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode models response: %w", err)
	}
	models := make([]ModelInfo, 0, len(result.Data))
	for _, item := range result.Data {
		// OpenAI 兼容接口未定义展示名称，直接使用模型 ID。
		models = append(models, ModelInfo{ID: item.ID, Name: item.ID})
	}
	return models, nil
}
