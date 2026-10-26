package chatcompletion

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/provider"
)

type Client struct {
	config provider.Config
}

func New(config provider.Config) *Client {
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	return &Client{config: config}
}

// Stream 串行发送协议分片，收到 [DONE] 后返回，不合并消息或工具调用。
func (c *Client) Stream(ctx context.Context, req Request, onChunk func(Chunk)) error {
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("encode completion request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.config.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create completion request: %w", err)
	}
	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	for name, values := range c.config.Headers {
		httpReq.Header.Del(name)
		for _, value := range values {
			httpReq.Header.Add(name, value)
		}
	}
	response, err := c.config.HTTPClient.Do(httpReq)
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
	return readStream(response.Body, onChunk)
}

func readStream(reader io.Reader, onChunk func(Chunk)) error {
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
			return nil
		}
		var chunk Chunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return fmt.Errorf("decode completion chunk: %w", err)
		}
		if chunk.Error != nil {
			return fmt.Errorf("completion error: %s", chunk.Error.Message)
		}
		onChunk(chunk)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read completion stream: %w", err)
	}
	return io.ErrUnexpectedEOF
}
