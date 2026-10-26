package chatcompletion

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/provider"
)

func TestProviderHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Custom token" {
			t.Errorf("unexpected authorization: %q", r.Header.Get("Authorization"))
		}
		if got := r.Header.Values("X-Provider"); !reflect.DeepEqual(got, []string{"first", "second"}) {
			t.Errorf("unexpected provider header: %v", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()
	client := New(provider.Config{
		BaseURL: server.URL,
		APIKey:  "default-key",
		Headers: http.Header{
			"Authorization": {"Custom token"},
			"X-Provider":    {"first", "second"},
		},
	})
	if err := client.Stream(context.Background(), Request{Model: "test-model"}, func(Chunk) {}); err != nil {
		t.Fatal(err)
	}
}
