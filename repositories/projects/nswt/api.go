package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type requestParams struct {
	URL     string
	Method  string
	Host    string
	Headers http.Header
	Body    []byte
}

type ciphercodeResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data *ciphercodeData `json:"data"`
}

type ciphercodeData struct {
	Ciphercode string `json:"ciphercode"`
	Safesalt   string `json:"safesalt"`
}

type asyncRushResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

var asyncRushFixedHeaders = http.Header{
	"Accept-Encoding": {"gzip, br"},
	"App-Version":     {"2.16.19"},
	"Connection":      {"keep-alive"},
	"Content-Type":    {"application/json; charset=utf-8"},
	"Referer":         {"https://servicewechat.com/wx9f30f1cea85e1e8c/1249/page-frame.html"},
	"User-Agent":      {"Mozilla/5.0 (iPhone; CPU iPhone OS 26_6_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.76(0x18004c39) NetType/WIFI Language/zh_CN"},
	"X-Page-Path":     {"pages/coupon/index?ref_src=index_35&id=d9qkmf566n467132aovg"},
	"X-Page-Uuid":     {"91fb5dab274451a20a1e3fe0101e7155"},
}

const (
	asyncRushHost = "nswtt.rim20.com"
	asyncRushURL  = "https://nswtt.rim20.com/api/wtt/coupon/rush/asyncrush"
)

func ciphercode(
	ctx context.Context,
	client *http.Client,
	params requestParams,
) (ciphercodeResponse, error) {
	var result ciphercodeResponse
	statusCode, status, err := callAPI(ctx, client, params, &result)
	if err != nil {
		return ciphercodeResponse{}, err
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return ciphercodeResponse{}, fmt.Errorf("unexpected HTTP status: %s", status)
	}
	return result, nil
}

func asyncRush(
	ctx context.Context,
	client *http.Client,
	params requestParams,
) (asyncRushResponse, error) {
	requestID, err := randomRequestID()
	if err != nil {
		return asyncRushResponse{}, fmt.Errorf("generate request ID: %w", err)
	}

	params.URL = asyncRushURL + "?requestid=" + requestID
	params.Method = http.MethodPost
	params.Host = asyncRushHost
	params.Headers = buildAsyncRushHeaders(params.Headers)

	var result asyncRushResponse
	statusCode, status, err := callAPI(ctx, client, params, &result)
	if err != nil {
		return asyncRushResponse{}, err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusBadRequest {
		return asyncRushResponse{}, fmt.Errorf("unexpected HTTP status: %s", status)
	}
	return result, nil
}

func buildAsyncRushHeaders(dynamic http.Header) http.Header {
	headers := asyncRushFixedHeaders.Clone()

	for name, values := range dynamic {
		lowerName := strings.ToLower(name)
		if lowerName != "cookie" && !strings.HasPrefix(lowerName, "x-app-") {
			continue
		}

		headers.Del(name)
		for _, value := range values {
			headers.Add(name, value)
		}
	}

	return headers
}

func callAPI(
	ctx context.Context,
	client *http.Client,
	params requestParams,
	result any,
) (int, string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := newAPIRequest(ctx, params)
	if err != nil {
		return 0, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("read response: %w", err)
	}
	if err := json.Unmarshal(body, result); err != nil {
		return 0, "", fmt.Errorf("decode response: %w", err)
	}

	return resp.StatusCode, resp.Status, nil
}

func newAPIRequest(ctx context.Context, params requestParams) (*http.Request, error) {
	method := params.Method
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		params.URL,
		bytes.NewReader(params.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header = params.Headers.Clone()
	req.Header.Del("Content-Length")
	if params.Host != "" {
		req.Host = params.Host
	}

	for name := range req.Header {
		if strings.EqualFold(name, "Host") {
			if req.Host == "" {
				req.Host = req.Header.Get(name)
			}
			req.Header.Del(name)
			break
		}
	}

	return req, nil
}
