package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestRunRushFlow(t *testing.T) {
	const cookie = "session=shared-cookie"

	callCount := 0
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			switch callCount {
			case 1:
				if got := req.Header.Get("Cookie"); got != cookie {
					t.Errorf("ciphercode Cookie = %q, want %q", got, cookie)
				}
				return jsonResponse(
					`{"code":0,"msg":"成功","data":{"ciphercode":"2279","safesalt":"iroO"}}`,
				), nil
			case 2:
				if got := req.Header.Get("Cookie"); got != cookie {
					t.Errorf("asyncRush Cookie = %q, want %q", got, cookie)
				}
				for _, name := range []string{
					"X-App-Sign",
					"X-App-Sn",
					"X-App-Timestamp",
				} {
					if req.Header.Get(name) == "" {
						t.Errorf("%s is empty", name)
					}
				}
				if req.URL.Query().Get("requestid") == "" {
					t.Error("requestid is empty")
				}
				return jsonResponse(
					`{"code":300001,"msg":"该场券已被抢光了～","data":null}`,
				), nil
			default:
				t.Fatalf("unexpected request #%d", callCount)
				return nil, nil
			}
		}),
	}

	cipherParams := requestParams{
		URL:    "https://example.test/api/wtt/coupon/rush/ciphercode",
		Method: http.MethodPost,
		Headers: http.Header{
			"Content-Type": {"application/json"},
			"Cookie":       {cookie},
		},
		Body: []byte(
			`{"batchid":"dag323566n47qqad5hug","batchcode":"zEWP3wnPkiGyRZE5cNSwanWwWHR6rY2Y2cDn","sign":"e6bd9440e0103343a32999e87c156f63"}`,
		),
	}
	result, err := runRushFlow(
		context.Background(),
		client,
		cipherParams,
		rushFlowOptions{},
	)
	if err != nil {
		t.Fatalf("runRushFlow() error = %v", err)
	}
	if callCount != 2 {
		t.Fatalf("request count = %d, want 2", callCount)
	}
	if result.Ciphercode.Data == nil || result.Ciphercode.Data.Ciphercode != "2279" {
		t.Fatalf("ciphercode result = %+v", result.Ciphercode)
	}
	if result.AsyncRush.Code != 300001 ||
		!strings.Contains(result.AsyncRush.Msg, "抢光") {
		t.Fatalf("async rush result = %+v", result.AsyncRush)
	}
}
