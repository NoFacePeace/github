package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Option func(req *http.Request)

func Do(ctx context.Context, url, method string, reqBody any, respBody any, options ...Option) error {
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("json marshal error : [%w]", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(raw))
	if err != nil {
		return fmt.Errorf("http new request error: [%w]", err)
	}
	for _, opt := range options {
		opt(req)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http default client do error: [%w]", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io read all error: [%w]", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("http response status code %d, body: %s", resp.StatusCode, string(body))
	}
	if respBody == nil {
		return nil
	}
	if err := json.Unmarshal(body, respBody); err != nil {
		return fmt.Errorf("json unmarshal error: [%w]", err)
	}
	return nil
}

func WithContentTypeJson() Option {
	return func(req *http.Request) {
		req.Header.Set("content-type", "application/json")
	}
}
