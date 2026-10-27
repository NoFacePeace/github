package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultProviderListModels(t *testing.T) {
	for _, authorization := range []string{"Bearer test-key", "Custom token"} {
		t.Run(authorization, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v1/models" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != authorization || r.Header.Get("X-Provider") != "example" {
					t.Errorf("unexpected request headers: %v", r.Header)
				}
				io.WriteString(w, `{"object":"list","data":[{"id":"model-a","owned_by":"example"},{"id":"model-b"}]}`)
			}))
			defer server.Close()
			config := Config{BaseURL: server.URL + "/v1/", APIKey: "test-key", Headers: http.Header{"X-Provider": {"example"}}, HTTPClient: server.Client()}
			if authorization != "Bearer test-key" {
				config.Headers.Set("Authorization", authorization)
			}
			models, err := New(config).ListModels(context.Background())
			want := []ModelInfo{{ID: "model-a", Name: "model-a"}, {ID: "model-b", Name: "model-b"}}
			if err != nil || !reflect.DeepEqual(models, want) {
				t.Fatalf("models = %+v, error = %v", models, err)
			}
		})
	}
}

func TestDefaultProviderErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"unauthorized", http.StatusUnauthorized, `{"error":"invalid key"}`, "401"},
		{"malformed response", http.StatusOK, "invalid", "decode models response"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				io.WriteString(w, tc.body)
			}))
			defer server.Close()
			_, err := New(Config{BaseURL: server.URL}).ListModels(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultProviderCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := New(Config{BaseURL: "http://localhost/v1"}).ListModels(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}
}
