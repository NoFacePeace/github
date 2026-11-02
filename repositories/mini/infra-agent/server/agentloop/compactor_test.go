package agentloop

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

func TestDedupResults(t *testing.T) {
	messages := []llm.Message{
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read", Arguments: []byte(`{"path":"a.go","limit":10,"__trace":"old"}`)},
		}},
		{Role: llm.RoleTool, ToolCallID: "call_1", Content: "old output"},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{
			{ID: "call_2", Name: "read", Arguments: []byte(`{ "limit": 10, "path": "a.go", "__trace": "new" }`)},
			{ID: "call_3", Name: "read", Arguments: []byte(`{"path":"b.go","limit":10}`)},
		}},
		{Role: llm.RoleTool, ToolCallID: "call_2", Content: "latest output"},
		{Role: llm.RoleTool, ToolCallID: "call_3", Content: "other output"},
	}
	got := dedupResults(messages)
	want := append([]llm.Message(nil), messages...)
	want[1].Content = "[dedup: read — same input, see latest result]"
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected messages: %#v", got)
	}
	if messages[1].Content != "old output" {
		t.Fatal("input messages were modified")
	}
}

func TestDedupResultsPreservesSpecialTools(t *testing.T) {
	for _, name := range []string{tools.NameAskUser, tools.NameGetSkillFile} {
		t.Run(name, func(t *testing.T) {
			messages := []llm.Message{
				{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "call_1", Name: name, Arguments: []byte(`{}`)}}},
				{Role: llm.RoleTool, ToolCallID: "call_1", Content: "first"},
				{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "call_2", Name: name, Arguments: []byte(`{}`)}}},
				{Role: llm.RoleTool, ToolCallID: "call_2", Content: "second"},
			}
			if got := dedupResults(messages); !reflect.DeepEqual(got, messages) {
				t.Fatalf("special tool results changed: %#v", got)
			}
		})
	}
}

func TestDedupPlans(t *testing.T) {
	plan := func(id string) llm.ToolCall {
		return llm.ToolCall{ID: id, Name: tools.NameUpdatePlan}
	}
	read := llm.ToolCall{ID: "read_1", Name: "read"}
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "goal"},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{plan("plan_1")}},
		{Role: llm.RoleTool, ToolCallID: "plan_1", Content: "old"},
		{Role: llm.RoleAssistant, Content: "progress", ToolCalls: []llm.ToolCall{plan("plan_2")}},
		{Role: llm.RoleTool, ToolCallID: "plan_2", Content: "old"},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{plan("plan_3"), read}},
		{Role: llm.RoleTool, ToolCallID: "plan_3", Content: "old"},
		{Role: llm.RoleTool, ToolCallID: "read_1", Content: "file"},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{plan("plan_4")}},
		{Role: llm.RoleTool, ToolCallID: "plan_4", Content: "latest"},
	}
	want := []llm.Message{
		messages[0],
		{Role: llm.RoleAssistant, Content: "progress", ToolCalls: []llm.ToolCall{}},
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{read}},
		messages[7], messages[8], messages[9],
	}
	if got := dedupPlans(messages); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected messages: %#v", got)
	}
	if len(messages[3].ToolCalls) != 1 || len(messages[5].ToolCalls) != 2 {
		t.Fatal("input tool calls were modified")
	}
	if got := dedupPlans(messages[7:]); !reflect.DeepEqual(got, messages[7:]) {
		t.Fatal("messages without duplicate plans changed")
	}
}

func TestTruncateResults(t *testing.T) {
	large := strings.Repeat("x", 20*1024+1)
	for _, tt := range []struct {
		name      string
		tool      string
		arguments string
		content   string
		count     int
	}{
		{name: "large output", tool: "read", content: large, count: 1},
		{name: "at threshold", tool: "read", content: large[:20*1024]},
		{name: "user answer", tool: tools.NameAskUser, content: large},
		{name: "skill reload", tool: tools.NameGetSkillFile, arguments: `{"skill_id":"coding","workspace_id":"ws_1","resource_path":"SKILL.md"}`, content: "instructions", count: 1},
		{name: "invalid skill arguments", tool: tools.NameGetSkillFile, arguments: `{`, content: "instructions", count: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			messages := []llm.Message{
				{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "old", Name: tt.tool, Arguments: []byte(tt.arguments)}}},
				{Role: llm.RoleTool, ToolCallID: "old", Content: tt.content},
			}
			for _, id := range []string{"recent_1", "recent_2", "recent_3"} {
				messages = append(messages,
					llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: id, Name: "read"}}},
					llm.Message{Role: llm.RoleTool, ToolCallID: id, Content: large})
			}
			got, count := truncateResults(messages)
			if count != tt.count || len(got) != len(messages) {
				t.Fatalf("count = %d, messages = %d", count, len(got))
			}
			if !reflect.DeepEqual(got[2:], messages[2:]) || messages[1].Content != tt.content {
				t.Fatal("recent results or input messages were modified")
			}
			if got[1].ToolCallID != "old" || got[1].Role != llm.RoleTool {
				t.Fatal("tool result pairing changed")
			}
			if tt.count == 0 {
				if got[1].Content != tt.content {
					t.Fatal("protected output changed")
				}
				return
			}
			var marker map[string]any
			if err := json.Unmarshal([]byte(got[1].Content), &marker); err != nil {
				t.Fatalf("invalid marker: %v", err)
			}
			if tt.name == "skill reload" {
				if marker["skill_id"] != "coding" || marker["workspace_id"] != "ws_1" || marker["action"] == nil {
					t.Fatalf("missing skill reload information: %#v", marker)
				}
			} else if marker["tool_call_id"] != "old" || marker["original_bytes"] != float64(len(tt.content)) {
				t.Fatalf("missing result information: %#v", marker)
			}
		})
	}
}

func TestOldestDroppableBlock(t *testing.T) {
	batch := func(id, name string) []llm.Message {
		return []llm.Message{
			{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: id, Name: name}}},
			{Role: llm.RoleTool, ToolCallID: id, Content: "output"},
		}
	}
	current := []llm.Message{{Role: llm.RoleSystem}, {Role: llm.RoleUser}, {Role: llm.RoleAssistant, Content: "progress"}}
	for _, id := range []string{"a", "b", "c", "d"} {
		current = append(current, batch(id, "read")...)
	}
	history := append([]llm.Message{{Role: llm.RoleSystem}}, batch("old", "read")...)
	history = append(history, llm.Message{Role: llm.RoleUser})
	pinned := append(batch("ask", tools.NameAskUser), llm.Message{Role: llm.RoleUser})
	incomplete := []llm.Message{
		{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "pending", Name: "read"}}},
		{Role: llm.RoleTool, ToolCallID: "orphan"},
		{Role: llm.RoleUser},
	}
	for _, tt := range []struct {
		name     string
		messages []llm.Message
		start    int
		end      int
		ok       bool
	}{
		{name: "empty"},
		{name: "historical complete pair", messages: history, start: 1, end: 3, ok: true},
		{name: "current old batch", messages: current, start: 3, end: 5, ok: true},
		{name: "recent batches protected", messages: current[:9]},
		{name: "special tool protected", messages: pinned},
		{name: "incomplete and orphan protected", messages: incomplete},
		{name: "latest user protected", messages: []llm.Message{{Role: llm.RoleSystem}, {Role: llm.RoleUser}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			start, end, ok := oldestDroppableBlock(tt.messages)
			if start != tt.start || end != tt.end || ok != tt.ok {
				t.Fatalf("got (%d, %d, %v), want (%d, %d, %v)", start, end, ok, tt.start, tt.end, tt.ok)
			}
		})
	}
}

func TestTrimMessagesDropsBlocksUntilTarget(t *testing.T) {
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "instructions"},
		{Role: llm.RoleUser, Content: strings.Repeat("a", 1000)},
		{Role: llm.RoleAssistant, Content: strings.Repeat("b", 1000)},
		{Role: llm.RoleUser, Content: "current"},
	}
	original := append([]llm.Message(nil), messages...)
	config := Config{TrimTriggerPercent: 80, TrimRetentionPercent: 50}
	got := trimMessages(messages, 10000, config)
	want := []llm.Message{messages[0], messages[3]}
	if !reflect.DeepEqual(got.Messages, want) || got.DroppedBlocks != 2 {
		t.Fatalf("unexpected trim result: %#v", got)
	}
	if got.AfterBytes > got.RetentionTargetBytes || got.BeforeBytes <= got.AfterBytes || got.TriggerThreshold != 8000 {
		t.Fatalf("unexpected trim statistics: %#v", got)
	}
	if !reflect.DeepEqual(messages, original) {
		t.Fatal("input messages were modified")
	}
}

func TestTrimMessagesStopsWhenAllBlocksProtected(t *testing.T) {
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "instructions"},
		{Role: llm.RoleUser, Content: strings.Repeat("a", 1000)},
	}
	got := trimMessages(messages, 10000, Config{TrimTriggerPercent: 80, TrimRetentionPercent: 50})
	if !reflect.DeepEqual(got.Messages, messages) || got.DroppedBlocks != 0 || got.AfterBytes <= got.RetentionTargetBytes {
		t.Fatalf("protected messages should remain above target: %#v", got)
	}
}

func TestNewDefaultsTrimConfig(t *testing.T) {
	runner := New(Config{}).(*loop)
	if runner.config.TrimRetentionPercent != 50 || runner.config.TrimTriggerPercent != 80 {
		t.Fatalf("unexpected defaults: %#v", runner.config)
	}
}

func TestToolInputKeyPreservesLargeNumbers(t *testing.T) {
	a := toolInputKey([]byte(`{"id":9007199254740992}`))
	b := toolInputKey([]byte(`{"id":9007199254740993}`))
	if a == b {
		t.Fatal("different large integers share the same key")
	}
}
