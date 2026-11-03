package tasks

import (
	"context"
	"errors"
	"testing"

	"infra-agent/server/agentloop"
	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

type authorizationStore struct {
	stubTaskStore
	tool    *ToolAuthorization
	network *NetworkAuthorization
	paths   []PathAuthorization
	err     error
	query   ToolAuthorizationQuery
}

func (s *authorizationStore) FindToolAuthorization(_ context.Context, query ToolAuthorizationQuery) (*ToolAuthorization, error) {
	s.query = query
	return s.tool, s.err
}

func (s *authorizationStore) GetNetworkAuthorization(context.Context, string) (*NetworkAuthorization, error) {
	return s.network, s.err
}

func (s *authorizationStore) ListPathAuthorizations(context.Context, string) ([]PathAuthorization, error) {
	return s.paths, s.err
}

func TestExecutorReusesAuthorizations(t *testing.T) {
	storageErr := errors.New("authorization storage unavailable")
	for _, tt := range []struct {
		name        string
		mode        ApprovalMode
		risk        tools.Risk
		store       authorizationStore
		wantWaiting bool
		wantErr     bool
	}{
		{name: "tool authorization", store: authorizationStore{tool: &ToolAuthorization{}}},
		{name: "missing tool authorization", wantWaiting: true},
		{name: "network does not replace tool authorization", risk: tools.Risk{RequiresNetwork: true}, store: authorizationStore{network: &NetworkAuthorization{RunID: "run_1"}}, wantWaiting: true},
		{name: "tool does not replace network authorization", risk: tools.Risk{RequiresNetwork: true}, store: authorizationStore{tool: &ToolAuthorization{}}, wantWaiting: true},
		{name: "auto reuses network authorization", mode: ApprovalModeAuto, risk: tools.Risk{RequiresNetwork: true}, store: authorizationStore{network: &NetworkAuthorization{RunID: "run_1"}}},
		{name: "other run network grant", mode: ApprovalModeAuto, risk: tools.Risk{RequiresNetwork: true}, store: authorizationStore{network: &NetworkAuthorization{RunID: "run_2"}}, wantWaiting: true},
		{name: "path read", risk: tools.Risk{Paths: []tools.PathAccess{{Path: "/data/docs/a", Operation: tools.PathRead}}}, store: authorizationStore{tool: &ToolAuthorization{}, paths: []PathAuthorization{{WorkspaceID: "workspace_1", Path: "/data/docs", Read: true}}}},
		{name: "read cannot authorize write", risk: tools.Risk{Paths: []tools.PathAccess{{Path: "/data/docs/a", Operation: tools.PathWrite}}}, store: authorizationStore{tool: &ToolAuthorization{}, paths: []PathAuthorization{{WorkspaceID: "workspace_1", Path: "/data/docs", Read: true}}}, wantWaiting: true},
		{name: "sibling path prefix", risk: tools.Risk{Paths: []tools.PathAccess{{Path: "/data/docs-other/a", Operation: tools.PathRead}}}, store: authorizationStore{tool: &ToolAuthorization{}, paths: []PathAuthorization{{WorkspaceID: "workspace_1", Path: "/data/docs", Read: true}}}, wantWaiting: true},
		{name: "parent traversal", risk: tools.Risk{Paths: []tools.PathAccess{{Path: "/data/docs/../secret", Operation: tools.PathRead}}}, store: authorizationStore{tool: &ToolAuthorization{}, paths: []PathAuthorization{{WorkspaceID: "workspace_1", Path: "/data/docs", Read: true}}}, wantWaiting: true},
		{name: "all paths must match", risk: tools.Risk{Paths: []tools.PathAccess{{Path: "/data/docs/a", Operation: tools.PathRead}, {Path: "/secret/a", Operation: tools.PathRead}}}, store: authorizationStore{tool: &ToolAuthorization{}, paths: []PathAuthorization{{WorkspaceID: "workspace_1", Path: "/data/docs", Read: true}}}, wantWaiting: true},
		{name: "other workspace", risk: tools.Risk{Paths: []tools.PathAccess{{Path: "/data/docs/a", Operation: tools.PathRead}}}, store: authorizationStore{tool: &ToolAuthorization{}, paths: []PathAuthorization{{WorkspaceID: "workspace_2", Path: "/data/docs", Read: true}}}, wantWaiting: true},
		{name: "storage failure", store: authorizationStore{err: storageErr}, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var entries []agentloop.EntryType
			e := &toolExecutor{
				safetyEvaluator: safetyEvaluatorFunc(allowSafety),
				approvalMode:    tt.mode, runID: "run_1", workspaceID: "workspace_1",
				store: &tt.store, lease: &lease{},
				tools: map[string]tools.Tool{"read": approvalRiskTool{approvalTool{namedTool("read"), tools.ApprovalRequired}, tt.risk}},
				recorder: toolRecorderFunc(func(_ context.Context, entry agentloop.Entry) error {
					entries = append(entries, entry.Type)
					return nil
				}),
			}
			results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "call_1", Name: "read", Arguments: []byte(`{}`)}})
			if tt.wantErr {
				if !errors.Is(err, storageErr) || len(entries) != 0 {
					t.Fatalf("err=%v, entries=%v", err, entries)
				}
				return
			}
			wantStatus := tools.StatusCompleted
			if tt.wantWaiting {
				wantStatus = tools.StatusWaitingForApproval
			}
			if err != nil || len(results) != 1 || results[0].Status != wantStatus {
				t.Fatalf("results=%v, err=%v", results, err)
			}
			if tt.mode != ApprovalModeAuto && tt.store.query != (ToolAuthorizationQuery{ToolName: "read", Key: "all", ToolCallID: "call_1", RunID: "run_1", WorkspaceID: "workspace_1"}) {
				t.Fatalf("unexpected query: %+v", tt.store.query)
			}
		})
	}
}
