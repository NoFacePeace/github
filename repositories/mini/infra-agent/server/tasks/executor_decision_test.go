package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"infra-agent/server/agentloop"
	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

type executingTool struct {
	namedTool
	execute func() (tools.Result, error)
}

type approvalRiskTool struct {
	approvalTool
	risk tools.Risk
}

func (t approvalRiskTool) Risk(context.Context, json.RawMessage, tools.Env) (tools.Risk, error) {
	return t.risk, nil
}

func TestExecutorApprovalMode(t *testing.T) {
	for _, tt := range []struct {
		name       string
		mode       ApprovalMode
		risk       tools.Risk
		action     DecisionAction
		wantStatus tools.Status
	}{
		{"default waits", "", tools.Risk{}, DecisionRequireApproval, tools.StatusWaitingForApproval},
		{"manual waits", ApprovalModeManual, tools.Risk{}, DecisionRequireApproval, tools.StatusWaitingForApproval},
		{"policy waits", ApprovalModePolicy, tools.Risk{}, DecisionRequireApproval, tools.StatusWaitingForApproval},
		{"auto approves", ApprovalModeAuto, tools.Risk{}, DecisionRequireApproval, tools.StatusCompleted},
		{"auto keeps network approval", ApprovalModeAuto, tools.Risk{RequiresNetwork: true}, DecisionRequireApproval, tools.StatusWaitingForApproval},
		{"auto preserves denial", ApprovalModeAuto, tools.Risk{}, DecisionDeny, tools.StatusCompleted},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var entries []agentloop.EntryType
			e := &toolExecutor{
				approvalMode: tt.mode,
				store:        &stubTaskStore{}, lease: &lease{},
				tools: map[string]tools.Tool{"read": approvalRiskTool{approvalTool{namedTool("read"), tools.ApprovalPolicy}, tt.risk}},
				safetyEvaluator: safetyEvaluatorFunc(func(context.Context, SafetyRequest) (Decision, error) {
					return Decision{Action: tt.action, Reason: "policy reason"}, nil
				}),
				recorder: toolRecorderFunc(func(_ context.Context, entry agentloop.Entry) error {
					entries = append(entries, entry.Type)
					return nil
				}),
			}
			results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "a", Name: "read", Arguments: []byte(`{}`)}})
			if err != nil || len(results) != 1 || results[0].Status != tt.wantStatus {
				t.Fatalf("results=%v, err=%v", results, err)
			}
			if tt.action == DecisionDeny && (len(entries) != 1 || entries[0] != agentloop.EntryTypeToolResult || !strings.Contains(results[0].Content, "denied")) {
				t.Fatalf("denial was overridden: results=%v, entries=%v", results, entries)
			}
		})
	}
}

func (t executingTool) Execute(context.Context, json.RawMessage, tools.Env) (tools.Result, error) {
	return t.execute()
}

func TestExecutorHandlesDecisions(t *testing.T) {
	storageErr := errors.New("storage unavailable")
	for _, tt := range []struct {
		name                                         string
		action                                       DecisionAction
		failEntry                                    agentloop.EntryType
		toolErr                                      error
		wantExecutions, wantResults, wantEvaluations int
		wantEntries                                  []agentloop.EntryType
	}{
		{name: "deny", action: DecisionDeny, wantResults: 2, wantEvaluations: 2, wantEntries: []agentloop.EntryType{agentloop.EntryTypeToolResult, agentloop.EntryTypeToolResult}},
		{name: "approval pauses batch", action: DecisionRequireApproval, wantResults: 1, wantEvaluations: 1, wantEntries: []agentloop.EntryType{agentloop.EntryTypeApprovalRequested}},
		{name: "allow", action: DecisionAllow, wantExecutions: 2, wantResults: 2, wantEvaluations: 2, wantEntries: []agentloop.EntryType{agentloop.EntryTypeToolStarted, agentloop.EntryTypeToolResult, agentloop.EntryTypeToolStarted, agentloop.EntryTypeToolResult}},
		{name: "start record failure", action: DecisionAllow, failEntry: agentloop.EntryTypeToolStarted, wantEvaluations: 1, wantEntries: []agentloop.EntryType{agentloop.EntryTypeToolStarted}},
		{name: "result record failure", action: DecisionAllow, failEntry: agentloop.EntryTypeToolResult, wantExecutions: 1, wantEvaluations: 1, wantEntries: []agentloop.EntryType{agentloop.EntryTypeToolStarted, agentloop.EntryTypeToolResult}},
		{name: "approval record failure", action: DecisionRequireApproval, failEntry: agentloop.EntryTypeApprovalRequested, wantEvaluations: 1, wantEntries: []agentloop.EntryType{agentloop.EntryTypeApprovalRequested}},
		{name: "tool error feedback", action: DecisionAllow, toolErr: errors.New("read failed"), wantExecutions: 2, wantResults: 2, wantEvaluations: 2, wantEntries: []agentloop.EntryType{agentloop.EntryTypeToolStarted, agentloop.EntryTypeToolResult, agentloop.EntryTypeToolStarted, agentloop.EntryTypeToolResult}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var entries []agentloop.EntryType
			executions, evaluations := 0, 0
			e := &toolExecutor{
				store: &stubTaskStore{}, lease: &lease{Owner: "executor_1", Epoch: 1},
				tools: map[string]tools.Tool{"read": executingTool{namedTool("read"), func() (tools.Result, error) {
					if len(entries) == 0 || entries[len(entries)-1] != agentloop.EntryTypeToolStarted {
						t.Error("tool executed before start record")
					}
					executions++
					return tools.Result{Status: tools.StatusCompleted, Content: "file content"}, tt.toolErr
				}}},
				safetyEvaluator: safetyEvaluatorFunc(func(context.Context, SafetyRequest) (Decision, error) {
					evaluations++
					return Decision{Action: tt.action, Reason: "policy reason"}, nil
				}),
				recorder: toolRecorderFunc(func(_ context.Context, entry agentloop.Entry) error {
					entries = append(entries, entry.Type)
					if entry.Type == tt.failEntry {
						return storageErr
					}
					return nil
				}),
			}
			results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "a", Name: "read", Arguments: []byte(`{}`)}, {ID: "b", Name: "read", Arguments: []byte(`{}`)}})
			if (tt.failEntry != "" && !errors.Is(err, storageErr)) || (tt.failEntry == "" && err != nil) {
				t.Fatalf("unexpected error: %v", err)
			}
			if executions != tt.wantExecutions || evaluations != tt.wantEvaluations || len(results) != tt.wantResults || !reflect.DeepEqual(entries, tt.wantEntries) {
				t.Fatalf("executions=%d, evaluations=%d, results=%v, entries=%v", executions, evaluations, results, entries)
			}
			if len(results) > 0 {
				if results[0].ToolCallID != "a" {
					t.Error("result lost call ID")
				}
				if tt.action == DecisionRequireApproval && results[0].Status != tools.StatusWaitingForApproval {
					t.Error("approval result not pending")
				}
				if tt.action == DecisionDeny && !strings.Contains(results[0].Content, "policy reason") {
					t.Error("denial reason missing")
				}
				if tt.toolErr != nil && !strings.Contains(results[0].Content, tt.toolErr.Error()) {
					t.Error("tool failure missing")
				}
			}
		})
	}
}

type invalidToolLeaseStore struct{ stubTaskStore }

func (*invalidToolLeaseStore) ValidateLease(context.Context, string, string, int64) (bool, error) {
	return false, nil
}

func TestExecutorStopsOnLostLease(t *testing.T) {
	e := &toolExecutor{
		store: &invalidToolLeaseStore{}, lease: &lease{},
		tools:           map[string]tools.Tool{"read": executingTool{namedTool("read"), func() (tools.Result, error) { t.Fatal("executed with lost lease"); return tools.Result{}, nil }}},
		safetyEvaluator: safetyEvaluatorFunc(allowSafety),
		recorder:        toolRecorderFunc(func(context.Context, agentloop.Entry) error { t.Fatal("recorded start with lost lease"); return nil }),
	}
	results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "a", Name: "read", Arguments: []byte(`{}`)}})
	if err == nil || len(results) != 0 {
		t.Fatalf("results=%v, err=%v", results, err)
	}
}
