package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"infra-agent/server/agentloop"
	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

type toolRecorderFunc func(context.Context, agentloop.Entry) error

type safetyEvaluatorFunc func(context.Context, SafetyRequest) (Decision, error)

func (f safetyEvaluatorFunc) Evaluate(ctx context.Context, request SafetyRequest) (Decision, error) {
	return f(ctx, request)
}

func allowSafety(context.Context, SafetyRequest) (Decision, error) {
	return Decision{Action: DecisionAllow}, nil
}

func (f toolRecorderFunc) Record(ctx context.Context, entry agentloop.Entry) error {
	return f(ctx, entry)
}

type executionRecordStore struct {
	stubTaskStore
	get func(context.Context, string, string) (ToolExecution, error)
}

func (s *executionRecordStore) GetToolExecution(ctx context.Context, runID, callID string) (ToolExecution, error) {
	return s.get(ctx, runID, callID)
}

type namedTool string

func (t namedTool) Name() string { return string(t) }

func (t namedTool) AuthorizationKey(json.RawMessage) (string, error) {
	return "all", nil
}

func (t namedTool) Execute(context.Context, json.RawMessage, tools.Env) (tools.Result, error) {
	return tools.Result{Status: tools.StatusCompleted, Content: "executed"}, nil
}

func (t namedTool) ApprovalRequirement() tools.ApprovalRequirement {
	return tools.ApprovalNone
}

func (t namedTool) Risk(context.Context, json.RawMessage, tools.Env) (tools.Risk, error) {
	return tools.Risk{}, nil
}

func (t namedTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object"}`)
}

type pathTool struct{ namedTool }

type approvalTool struct {
	namedTool
	requirement tools.ApprovalRequirement
}

func (t approvalTool) ApprovalRequirement() tools.ApprovalRequirement { return t.requirement }

func TestExecutorEvaluatesSafetyByRequirement(t *testing.T) {
	for _, requirement := range []tools.ApprovalRequirement{tools.ApprovalNone, tools.ApprovalPolicy, tools.ApprovalRequired} {
		t.Run(string(requirement), func(t *testing.T) {
			count := 0
			wantErr := errors.New("policy unavailable")
			ctx := context.Background()
			e := &toolExecutor{
				tools:    map[string]tools.Tool{"read": approvalTool{namedTool("read"), requirement}},
				store:    &stubTaskStore{},
				recorder: toolRecorderFunc(func(context.Context, agentloop.Entry) error { return nil }),
				safetyEvaluator: safetyEvaluatorFunc(func(got context.Context, request SafetyRequest) (Decision, error) {
					count++
					if got != ctx || request.ToolName != "read" {
						t.Error("unexpected safety request")
					}
					return Decision{}, wantErr
				}),
			}
			_, err := e.Execute(ctx, []llm.ToolCall{{ID: "a", Name: "read", Arguments: []byte(`{}`)}})
			if !errors.Is(err, wantErr) || count != 1 {
				t.Fatalf("evaluations=%d, err=%v", count, err)
			}
		})
	}
}

func (t pathTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","required":["path"],"properties":{"path":{"type":"string"}}}`)
}

func TestExecutorRecordsSchemaFailure(t *testing.T) {
	for _, failedRecord := range []bool{false, true} {
		wantErr := errors.New("storage unavailable")
		records := 0
		e := &toolExecutor{
			tools: map[string]tools.Tool{"read": pathTool{namedTool("read")}},
			store: &stubTaskStore{},
			recorder: toolRecorderFunc(func(_ context.Context, entry agentloop.Entry) error {
				records++
				if entry.Type != agentloop.EntryTypeToolResult || entry.Data["tool_call_id"] != "a" {
					t.Fatalf("unexpected entry: %#v", entry)
				}
				if failedRecord {
					return wantErr
				}
				return nil
			}),
		}
		results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "a", Name: "read", Arguments: []byte(`{"path":123}`)}})
		if records != 1 {
			t.Fatalf("record count=%d", records)
		}
		if failedRecord {
			if !errors.Is(err, wantErr) || len(results) != 0 {
				t.Fatalf("results=%v, err=%v", results, err)
			}
		} else if err != nil || len(results) != 1 || !strings.Contains(results[0].Content, "/path") {
			t.Fatalf("results=%v, err=%v", results, err)
		}
	}
}

func TestExecutorChecksExecutionRecord(t *testing.T) {
	stored := tools.Result{ToolCallID: "a", Status: tools.StatusCompleted, Content: "saved output"}
	readErr := errors.New("log corrupted")
	recordErr := errors.New("write failed")
	for _, tt := range []struct {
		name        string
		state       ToolExecution
		readErr     error
		recordErr   error
		wantErr     error
		wantRecords int
		wantResults int
	}{
		{name: "existing result", state: ToolExecution{Started: true, Result: &stored}, wantResults: 1},
		{name: "unknown result", state: ToolExecution{Started: true}, wantRecords: 1, wantResults: 1},
		{name: "not started", wantRecords: 2, wantResults: 1},
		{name: "read failure", readErr: readErr, wantErr: readErr},
		{name: "record failure", state: ToolExecution{Started: true}, recordErr: recordErr, wantErr: recordErr, wantRecords: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			reads, records := 0, 0
			e := &toolExecutor{
				runID:           "run_1",
				lease:           &lease{Owner: "executor_1", Epoch: 1},
				safetyEvaluator: safetyEvaluatorFunc(allowSafety),
				tools:           map[string]tools.Tool{"read": namedTool("read")},
				store: &executionRecordStore{get: func(got context.Context, runID, callID string) (ToolExecution, error) {
					reads++
					if got != ctx || runID != "run_1" || callID != "a" {
						t.Error("incorrect execution lookup")
					}
					return tt.state, tt.readErr
				}},
				recorder: toolRecorderFunc(func(_ context.Context, entry agentloop.Entry) error {
					records++
					if (entry.Type != agentloop.EntryTypeToolResult && entry.Type != agentloop.EntryTypeToolStarted) || entry.Data["tool_call_id"] != "a" {
						t.Error("incorrect recovery result")
					}
					return tt.recordErr
				}),
			}
			results, err := e.Execute(ctx, []llm.ToolCall{{ID: "a", Name: "read", Arguments: []byte(`{}`)}})
			if !errors.Is(err, tt.wantErr) || reads != 1 || records != tt.wantRecords || len(results) != tt.wantResults {
				t.Fatalf("results=%v, err=%v, reads=%d, records=%d", results, err, reads, records)
			}
			if tt.state.Result != nil && (len(results) != 1 || results[0] != stored) {
				t.Fatal("stored result was not reused")
			}
		})
	}
}

func TestExecutorRecordsUnknownTool(t *testing.T) {
	var entries []agentloop.Entry
	e := &toolExecutor{recorder: toolRecorderFunc(func(_ context.Context, entry agentloop.Entry) error {
		entries = append(entries, entry)
		return nil
	})}
	results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "a", Name: "missing_a"}, {ID: "b", Name: "missing_b"}})
	if err != nil || len(results) != 2 || len(entries) != 2 {
		t.Fatalf("results=%v, entries=%v, err=%v", results, entries, err)
	}
	for i, result := range results {
		entry := entries[i]
		if entry.Type != agentloop.EntryTypeToolResult || entry.Data["tool_call_id"] != result.ToolCallID || entry.Data["content"] != result.Content || result.Status != tools.StatusCompleted {
			t.Fatalf("entry does not match returned result: %#v, %#v", entry, result)
		}
	}
}

func TestExecutorStopsOnRecordFailure(t *testing.T) {
	wantErr := errors.New("storage unavailable")
	count := 0
	e := &toolExecutor{recorder: toolRecorderFunc(func(context.Context, agentloop.Entry) error {
		count++
		if count == 2 {
			return wantErr
		}
		return nil
	})}
	results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "a"}, {ID: "b"}, {ID: "c"}})
	if !errors.Is(err, wantErr) || count != 2 || len(results) != 1 || results[0].ToolCallID != "a" {
		t.Fatalf("results=%v, records=%d, err=%v", results, count, err)
	}
}

func TestExecutorValidatesArgumentsJSON(t *testing.T) {
	for _, tt := range []struct {
		name      string
		arguments string
		invalid   bool
	}{
		{name: "empty", invalid: true},
		{name: "malformed", arguments: `{"path":`, invalid: true},
		{name: "trailing JSON", arguments: `{} {}`, invalid: true},
		{name: "valid", arguments: `{"path":"a.go"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			records := 0
			e := &toolExecutor{
				safetyEvaluator: safetyEvaluatorFunc(allowSafety),
				lease:           &lease{Owner: "executor_1", Epoch: 1},
				tools:           map[string]tools.Tool{"read": namedTool("read")},
				store:           &stubTaskStore{},
				recorder: toolRecorderFunc(func(_ context.Context, entry agentloop.Entry) error {
					records++
					if tt.invalid && (entry.Type != agentloop.EntryTypeToolResult || entry.Data["tool_call_id"] != "a" || entry.Data["content"] != "tool call arguments must be valid JSON") {
						t.Fatalf("unexpected invalid arguments entry: %#v", entry)
					}
					return nil
				}),
			}
			results, err := e.Execute(context.Background(), []llm.ToolCall{{ID: "a", Name: "read", Arguments: []byte(tt.arguments)}})
			want := 2
			if tt.invalid {
				want = 1
			}
			if err != nil || records != want || len(results) != 1 {
				t.Fatalf("results=%v, records=%d, err=%v", results, records, err)
			}
		})
	}
}
