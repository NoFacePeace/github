package tasks

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"infra-agent/server/providers"
	"infra-agent/server/workspaces"
)

type stubModelManager struct {
	ctx    context.Context
	calls  int
	models []providers.Model
	err    error
}

func (m *stubModelManager) List(ctx context.Context) ([]providers.Model, error) {
	m.ctx = ctx
	m.calls++
	return m.models, m.err
}

type stubProviderManager struct {
	ctx       context.Context
	ids       []string
	providers map[string]providers.Provider
	err       error
}

func (m *stubProviderManager) Get(ctx context.Context, id string) (providers.Provider, error) {
	m.ctx = ctx
	m.ids = append(m.ids, id)
	if m.err != nil {
		return providers.Provider{}, m.err
	}
	provider, ok := m.providers[id]
	if !ok {
		return providers.Provider{}, errors.New("provider not found")
	}
	return provider, nil
}

func TestResolveModelMaxOutputTokens(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		models  []providers.Model
		want    int
		wantIDs []string
		wantErr string
	}{
		{
			name:  "filters and matches provider name",
			model: "provider/model",
			models: []providers.Model{
				{ProviderID: "disabled", ModelID: "model", MaxOutputTokens: 100},
				{ProviderID: "unrelated", ModelID: "other-model", Enabled: true, MaxOutputTokens: 200},
				{ProviderID: "provider_2", ModelID: "model", Enabled: true, MaxOutputTokens: 400},
				{ID: "record_1", ProviderID: "provider_1", Name: "display name", ModelID: "model", Enabled: true, MaxOutputTokens: 8192},
				{ProviderID: "unused", ModelID: "model", Enabled: true, MaxOutputTokens: 16000},
			},
			want:    8192,
			wantIDs: []string{"provider_2", "provider_1"},
		},
		{
			name:    "model id contains slash",
			model:   "provider/namespace/model",
			models:  []providers.Model{{ProviderID: "provider_1", ModelID: "namespace/model", Enabled: true, MaxOutputTokens: 4096}},
			want:    4096,
			wantIDs: []string{"provider_1"},
		},
		{name: "empty list", model: "provider/model", wantErr: `model "provider/model" not found or disabled`},
		{
			name:    "disabled model",
			model:   "provider/model",
			models:  []providers.Model{{ProviderID: "disabled", ModelID: "model", MaxOutputTokens: 8192}},
			wantErr: `model "provider/model" not found or disabled`,
		},
		{
			name:    "model id mismatch",
			model:   "provider/model",
			models:  []providers.Model{{ProviderID: "unrelated", ModelID: "other-model", Enabled: true, MaxOutputTokens: 8192}},
			wantErr: `model "provider/model" not found or disabled`,
		},
		{
			name:    "provider name mismatch",
			model:   "provider/model",
			models:  []providers.Model{{ProviderID: "provider_2", ModelID: "model", Enabled: true, MaxOutputTokens: 8192}},
			wantIDs: []string{"provider_2"},
			wantErr: `model "provider/model" not found or disabled`,
		},
		{
			name:    "output limit not configured",
			model:   "provider/model",
			models:  []providers.Model{{ProviderID: "provider_1", ModelID: "model", Enabled: true}},
			wantIDs: []string{"provider_1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			modelManager := &stubModelManager{models: tt.models}
			providerManager := &stubProviderManager{providers: map[string]providers.Provider{
				"provider_1": {ID: "provider_1", Name: "provider"},
				"provider_2": {ID: "provider_2", Name: "other-provider"},
			}}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, modelManager, providerManager).(*service)

			got, err := taskService.resolveModelMaxOutputTokens(ctx, tt.model)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("resolveModelMaxOutputTokens() error = %v", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("resolveModelMaxOutputTokens() error = %v, want %q", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("resolveModelMaxOutputTokens() = %d, want %d", got, tt.want)
			}
			if modelManager.calls != 1 || modelManager.ctx != ctx {
				t.Error("List() must be called once with the request context")
			}
			if !slices.Equal(providerManager.ids, tt.wantIDs) {
				t.Errorf("Get() ids = %v, want %v", providerManager.ids, tt.wantIDs)
			}
			if len(tt.wantIDs) > 0 && providerManager.ctx != ctx {
				t.Error("Get() did not receive the request context")
			}
		})
	}
}

func TestResolveModelMaxOutputTokensRejectsInvalidIdentifier(t *testing.T) {
	for _, model := range []string{"", "model", "/model", "provider/", "/"} {
		t.Run(model, func(t *testing.T) {
			modelManager := &stubModelManager{}
			providerManager := &stubProviderManager{}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, modelManager, providerManager).(*service)

			got, err := taskService.resolveModelMaxOutputTokens(context.Background(), model)
			if err == nil || !strings.HasPrefix(err.Error(), "invalid model identifier ") {
				t.Fatalf("resolveModelMaxOutputTokens() error = %v, want invalid identifier error", err)
			}
			if got != 0 {
				t.Errorf("resolveModelMaxOutputTokens() = %d, want 0", got)
			}
			if modelManager.calls != 0 || len(providerManager.ids) != 0 {
				t.Error("invalid identifiers must not query managers")
			}
		})
	}
}

func TestResolveModelMaxOutputTokensWrapsErrors(t *testing.T) {
	wantErr := errors.New("storage unavailable")
	tests := []struct {
		name         string
		listErr      error
		getErr       error
		wantMessage  string
		wantGetCalls int
	}{
		{name: "list", listErr: wantErr, wantMessage: "list models: storage unavailable"},
		{name: "get provider", getErr: wantErr, wantMessage: `get provider "provider_1": storage unavailable`, wantGetCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelManager := &stubModelManager{
				models: []providers.Model{{ProviderID: "provider_1", ModelID: "model", Enabled: true, MaxOutputTokens: 8192}},
				err:    tt.listErr,
			}
			providerManager := &stubProviderManager{err: tt.getErr}
			taskService := NewTaskService(&stubWorkspaceService{}, workspaces.Config{}, Config{}, modelManager, providerManager).(*service)

			got, err := taskService.resolveModelMaxOutputTokens(context.Background(), "provider/model")
			if !errors.Is(err, wantErr) {
				t.Fatalf("resolveModelMaxOutputTokens() error = %v, want wrapped %v", err, wantErr)
			}
			if err.Error() != tt.wantMessage {
				t.Errorf("error message = %q, want %q", err.Error(), tt.wantMessage)
			}
			if got != 0 {
				t.Errorf("resolveModelMaxOutputTokens() = %d, want 0", got)
			}
			if len(providerManager.ids) != tt.wantGetCalls {
				t.Errorf("Get() calls = %d, want %d", len(providerManager.ids), tt.wantGetCalls)
			}
		})
	}
}
