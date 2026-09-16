package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunRushFlow(t *testing.T) {
	const cookie = "session=shared-cookie"

	var callCount atomic.Int32
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callNumber := int(callCount.Add(1))
			switch callNumber {
			case 1, 2, 3, 4, 5:
				if got := req.Header.Get("Cookie"); got != cookie {
					t.Errorf("ciphercode Cookie = %q, want %q", got, cookie)
				}
				return jsonResponse(
					`{"code":300010,"msg":"抢券未开始","data":null}`,
				), nil
			case 6:
				return jsonResponse(
					`{"code":0,"msg":"成功","data":{"ciphercode":"2279","safesalt":"iroO"}}`,
				), nil
			case 7:
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
				t.Fatalf("unexpected request #%d", callNumber)
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
	result, err := runRushFlowNow(
		context.Background(),
		client,
		cipherParams,
		rushFlowOptions{ReleaseTime: time.Now().Add(time.Second)},
	)
	if err != nil {
		t.Fatalf("runRushFlow() error = %v", err)
	}
	if got := callCount.Load(); got != 7 {
		t.Fatalf("request count = %d, want 7", got)
	}
	if result.Ciphercode.Data == nil || result.Ciphercode.Data.Ciphercode != "2279" {
		t.Fatalf("ciphercode result = %+v", result.Ciphercode)
	}
	if result.AsyncRush.Code != 300001 ||
		!strings.Contains(result.AsyncRush.Msg, "抢光") {
		t.Fatalf("async rush result = %+v", result.AsyncRush)
	}
}

func TestRushReleaseTime(t *testing.T) {
	now := time.Date(2026, time.September, 13, 9, 30, 0, 0, shanghaiLocation)
	got := rushReleaseTime(now)
	want := time.Date(2026, time.September, 13, 12, 0, 0, 0, shanghaiLocation)

	if !got.Equal(want) {
		t.Fatalf("rushReleaseTime() = %s, want %s", got, want)
	}
}

func TestCiphercodeSerialRetry(t *testing.T) {
	for _, tc := range []struct {
		name    string
		body    string
		expired bool
		cancel  bool
		want    int
	}{
		{"limit", `{"code":300010,"data":null}`, false, false, 6},
		{"expired", `{"code":300010,"data":null}`, true, false, 1},
		{"other error", `{"code":300006,"data":null}`, false, false, 1},
		{"success", `{"code":0,"data":{"ciphercode":"test","safesalt":"test"}}`, false, false, 1},
		{"cancel", `{"code":300010,"data":null}`, false, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if tc.cancel {
					cancel()
				}
				return jsonResponse(tc.body), nil
			})}
			deadline := time.Now().Add(time.Second)
			if tc.expired {
				deadline = time.Now().Add(-time.Second)
			}
			_, err := requestCiphercodeUntilOpen(ctx, client, requestParams{
				URL: "https://example.test/ciphercode",
			}, deadline)
			if tc.cancel {
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error = %v, want cancellation", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if calls != tc.want {
				t.Fatalf("calls = %d, want %d", calls, tc.want)
			}
		})
	}
}
