package watch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"umo/controller/healthcheck"
)

// WatchClient 定义 Watch 客户端的运行接口。
type WatchClient interface {
	// Run 在连接结束后重新连接，直到上下文取消。
	Run(ctx context.Context)
}

type client struct {
	url         string
	submitEvent func(context.Context, healthcheck.Event) error
}

// NewClient 创建 Watch 客户端。
func NewClient(url string, submitEvent func(context.Context, healthcheck.Event) error) WatchClient {
	return client{url: url, submitEvent: submitEvent}
}

func (c client) Run(ctx context.Context) {
	if c.url == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		log.Printf("create watch request: %v", err)
		return
	}

	const maxRetryDelay = 30 * time.Second
	retryDelay := time.Second
	for {
		if c.connect(req.Clone(ctx)) {
			retryDelay = time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(retryDelay):
		}
		retryDelay = min(retryDelay*2, maxRetryDelay)
	}
}

// connect 建立一次连接，返回是否成功读到过事件。
func (c client) connect(req *http.Request) bool {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Printf("connect to watch endpoint: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("watch endpoint returned HTTP %d", resp.StatusCode)
		return false
	}

	const heartbeatTimeout = 30 * time.Second
	heartbeatTimer := time.AfterFunc(heartbeatTimeout, cancel)
	defer heartbeatTimer.Stop()

	decoder := json.NewDecoder(resp.Body)
	receivedEvent := false
	var managementOffset uint64
	for {
		var event Event
		if err := decoder.Decode(&event); err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("decode watch event: %v", err)
			}
			return receivedEvent
		}
		receivedEvent = true
		if event.Type == EventTypeHeartbeat {
			heartbeatTimer.Reset(heartbeatTimeout)
			continue
		}
		offset := event.Offset.ManagementOffset
		if event.Type == EventTypeManagement && offset <= managementOffset {
			continue
		}
		if err := c.handleEvent(ctx, event); err != nil {
			log.Printf("handle watch event %q: %v", event.Type, err)
			continue
		}
		if event.Type == EventTypeFullSync || event.Type == EventTypeManagement {
			managementOffset = offset
		}
	}
}
