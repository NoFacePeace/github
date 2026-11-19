package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

type webhookRoundTripFunc func(*http.Request) (*http.Response, error)

func (f webhookRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// 验证慢响应不阻塞本轮处理，且通知保留冷却跳过和写入失败的 Pod。
func TestProcessStatusesWebhook(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 4*time.Second)
	defer cancel()
	received := make(chan webhookPayload, 1)
	release := make(chan struct{})
	transportDone := make(chan struct{})

	created, err := NewManager(nil, Config{Policies: []Policy{{
		Action: ActionMigrate, Status: []Status{StatusNodeNotReady},
		ReconcileInterval: time.Minute, Threshold: 1, Webhooks: []string{"https://example.test/webhook"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	m := created.(*manager)
	m.webhookClient = &http.Client{Transport: webhookRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		defer close(transportDone)
		defer r.Body.Close()
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected webhook request: %s %s", r.Method, r.Header.Get("Content-Type"))
		}
		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			return nil, err
		}
		received <- payload
		select {
		case <-release:
			return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	})}
	var pods []client.Object
	statuses := make(map[string][]Status)
	for _, name := range []string{"written", "cooldown", "failed"} {
		pod := &corev1.Pod{}
		pod.Namespace, pod.Name = "default", name
		pods = append(pods, pod)
		statuses[name] = []Status{StatusNodeNotReady}
	}
	m.k8sClient = fake.NewClientBuilder().WithObjects(pods...).WithInterceptorFuncs(interceptor.Funcs{
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			if obj.GetName() == "failed" {
				return errors.New("patch unavailable")
			}
			return c.Patch(ctx, obj, patch, opts...)
		},
	}).Build()
	m.podActionExecutedAt["cooldown"] = map[Action]time.Time{ActionMigrate: time.Now()}

	done := make(chan struct{})
	go func() {
		m.processStatuses(ctx, statuses)
		close(done)
	}()
	select {
	case payload := <-received:
		got := make(map[string]string)
		for _, pod := range payload.Pods {
			if pod.Namespace != "default" {
				t.Errorf("unexpected namespace for %q: %q", pod.Name, pod.Namespace)
			}
			got[pod.Name] = pod.Result
		}
		want := map[string]string{"written": "label_written", "cooldown": "skipped_cooldown", "failed": "label_write_failed"}
		if payload.Action != ActionMigrate || len(payload.Pods) != len(want) || !reflect.DeepEqual(got, want) {
			t.Errorf("unexpected webhook payload: %+v", payload)
		}
	case <-ctx.Done():
		t.Fatal("webhook was not received")
	}
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("processing blocked on webhook response")
	}
	close(release)
	<-transportDone
}
