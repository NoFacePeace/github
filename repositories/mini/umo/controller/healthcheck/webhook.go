package healthcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

type webhookPayload struct {
	Action Action             `json:"action"`
	Pods   []webhookPodResult `json:"pods"`
}

type webhookPodResult struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Result    string `json:"result"`
}

var defaultWebhookClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// postWebhooks 向策略配置的地址发送本轮结果。
func (m *manager) postWebhooks(ctx context.Context, urls []string, payload webhookPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("healthcheck webhook encode failed for action %q: %v", payload.Action, err)
		return
	}
	for i, url := range urls {
		if ctx.Err() != nil {
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			log.Printf("healthcheck webhook %d for action %q has an invalid URL", i+1, payload.Action)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := m.webhookClient.Do(req)
		if err != nil {
			log.Printf("healthcheck webhook %d for action %q failed: %v", i+1, payload.Action, errors.Unwrap(err))
			continue
		}
		_, _ = io.CopyN(io.Discard, resp.Body, 4096)
		resp.Body.Close()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			log.Printf("healthcheck webhook %d for action %q returned HTTP %d", i+1, payload.Action, resp.StatusCode)
		}
	}
}
