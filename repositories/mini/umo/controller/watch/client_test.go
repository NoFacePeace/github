package watch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"umo/controller/healthcheck"
)

type watchRoundTripFunc func(*http.Request) (*http.Response, error)

func (f watchRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// 验证全量偏移量用于增量去重，入队失败时仍可接收相同偏移量。
func TestManagementOffsetDeduplication(t *testing.T) {
	const stream = `{"eventType":"fullSync","data":"{}","eventOffset":{"managementOffset":100}}
{"eventType":"management","data":"{\"cluster_nodes\":{\"pod-a\":\"snapshot_duplicate\"}}","eventOffset":{"managementOffset":100}}
{"eventType":"management","data":"{\"cluster_nodes\":{\"pod-a\":\"first_attempt\"}}","eventOffset":{"managementOffset":105}}
{"eventType":"management","data":"{\"cluster_nodes\":{\"pod-a\":\"retry\"}}","eventOffset":{"managementOffset":105}}
{"eventType":"management","data":"{\"cluster_nodes\":{\"pod-a\":\"duplicate\"}}","eventOffset":{"managementOffset":105}}`
	original := http.DefaultClient
	defer func() { http.DefaultClient = original }()
	http.DefaultClient = &http.Client{Transport: watchRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(stream))}, nil
	})}
	var submitted []healthcheck.Status
	c := client{submitEvent: func(_ context.Context, event healthcheck.Event) error {
		status := event.ClusterNodes["pod-a"]
		submitted = append(submitted, status)
		if status == "first_attempt" {
			return errors.New("submit failed")
		}
		return nil
	}}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.test/watch", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.connect(req)
	if want := []healthcheck.Status{"first_attempt", "retry"}; !slices.Equal(submitted, want) {
		t.Fatalf("submitted %v, want %v", submitted, want)
	}
}

// 验证心跳延长连接存活时间，超时后客户端重新连接。
func TestRunReconnectsAfterHeartbeatTimeout(t *testing.T) {
	original := http.DefaultClient
	defer func() { http.DefaultClient = original }()
	synctest.Test(t, func(t *testing.T) {
		reader, writer := io.Pipe()
		attempts := 0
		http.DefaultClient = &http.Client{Transport: watchRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts > 1 {
				<-req.Context().Done()
				return nil, req.Context().Err()
			}
			go func() {
				<-req.Context().Done()
				writer.CloseWithError(req.Context().Err())
			}()
			return &http.Response{StatusCode: http.StatusOK, Body: reader}, nil
		})}

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() {
			defer close(done)
			NewClient("https://example.test/watch", nil).Run(ctx)
		}()
		defer func() {
			cancel()
			<-done
		}()
		synctest.Wait()

		time.Sleep(20 * time.Second)
		if _, err := io.WriteString(writer, `{"eventType":"heartbeat"}`); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		time.Sleep(12 * time.Second)
		synctest.Wait()
		if attempts != 1 {
			t.Fatalf("connection attempts after heartbeat = %d, want 1", attempts)
		}
		time.Sleep(20 * time.Second)
		synctest.Wait()
		if attempts != 2 {
			t.Fatalf("connection attempts after heartbeat timeout = %d, want 2", attempts)
		}
	})
}
