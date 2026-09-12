package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestCiphercode(t *testing.T) {
	client := responseClient(t, `{"code":0,"msg":"成功","data":{"ciphercode":"2279","safesalt":"iroO"}}`)
	params := testRequestParams()

	got, err := ciphercode(context.Background(), client, params)
	if err != nil {
		t.Fatalf("ciphercode() error = %v", err)
	}
	if got.Code != 0 || got.Data == nil {
		t.Fatalf("ciphercode() = %+v", got)
	}
	if got.Data.Ciphercode != "2279" || got.Data.Safesalt != "iroO" {
		t.Fatalf("ciphercode() data = %+v", got.Data)
	}
}

func TestAsyncRush(t *testing.T) {
	params := testRequestParams()
	params.Headers.Set("X-App-Sign", "dynamic-sign")
	params.Headers.Set("X-App-Sn", "dynamic-sn")
	params.Headers.Set("X-App-Timestamp", "dynamic-timestamp")
	params.Headers.Set("User-Agent", "should-not-be-used")

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path; got != asyncRushURL {
				t.Errorf("URL = %q, want %q", got, asyncRushURL)
			}
			requestID := req.URL.Query().Get("requestid")
			rawRequestID, err := base64.RawURLEncoding.DecodeString(requestID)
			if err != nil {
				t.Errorf("requestid %q is not URL-safe Base64: %v", requestID, err)
			}
			if len(requestID) != 32 || len(rawRequestID) != 24 {
				t.Errorf("requestid length = %d, decoded length = %d", len(requestID), len(rawRequestID))
			}
			if req.Host != asyncRushHost {
				t.Errorf("host = %q, want %q", req.Host, asyncRushHost)
			}
			if got := req.Header.Get("Cookie"); got != "session=test" {
				t.Errorf("Cookie = %q, want session=test", got)
			}
			if got := req.Header.Get("X-App-Sign"); got != "dynamic-sign" {
				t.Errorf("X-App-Sign = %q", got)
			}
			if got := req.Header.Get("X-App-Sn"); got != "dynamic-sn" {
				t.Errorf("X-App-Sn = %q", got)
			}
			if got := req.Header.Get("X-App-Timestamp"); got != "dynamic-timestamp" {
				t.Errorf("X-App-Timestamp = %q", got)
			}
			if got := req.Header.Get("User-Agent"); got != asyncRushFixedHeaders.Get("User-Agent") {
				t.Errorf("User-Agent = %q", got)
			}

			return jsonResponseWithStatus(
				http.StatusBadRequest,
				`{"code":300001,"msg":"该场券已被抢光了～","data":null}`,
			), nil
		}),
	}

	got, err := asyncRush(context.Background(), client, params)
	if err != nil {
		t.Fatalf("asyncRush() error = %v", err)
	}
	if got.Code != 300001 || got.Msg != "该场券已被抢光了～" {
		t.Fatalf("asyncRush() = %+v", got)
	}
}

func TestFormatCiphercodeResponse(t *testing.T) {
	result := ciphercodeResponse{
		Code: 0,
		Msg:  "成功",
		Data: &ciphercodeData{
			Ciphercode: "2279",
			Safesalt:   "iroO",
		},
	}

	got := formatCiphercodeResponse(result)
	want := "获取口令成功: 成功\nciphercode: 2279\nsafesalt: iroO"
	if got != want {
		t.Fatalf("formatCiphercodeResponse() = %q, want %q", got, want)
	}
}

func testRequestParams() requestParams {
	return requestParams{
		URL:    "https://example.test/api",
		Method: http.MethodPost,
		Host:   "example.test",
		Headers: http.Header{
			"Content-Type": {"application/json"},
			"Cookie":       {"session=test"},
		},
		Body: []byte("encrypted-request-body"),
	}
}

func responseClient(t *testing.T, responseBody string) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", req.Method)
			}
			if req.Host != "example.test" {
				t.Errorf("host = %q, want example.test", req.Host)
			}
			if got := req.Header.Get("Cookie"); got != "session=test" {
				t.Errorf("Cookie = %q, want session=test", got)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if got := string(body); got != "encrypted-request-body" {
				t.Errorf("body = %q, want encrypted-request-body", got)
			}

			return jsonResponse(responseBody), nil
		}),
	}
}

func jsonResponse(body string) *http.Response {
	return jsonResponseWithStatus(http.StatusOK, body)
}

func jsonResponseWithStatus(statusCode int, body string) *http.Response {
	return &http.Response{
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
