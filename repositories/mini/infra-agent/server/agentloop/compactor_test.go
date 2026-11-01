package agentloop

import (
	"reflect"
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

func TestToolInputKeyPreservesLargeNumbers(t *testing.T) {
	a := toolInputKey([]byte(`{"id":9007199254740992}`))
	b := toolInputKey([]byte(`{"id":9007199254740993}`))
	if a == b {
		t.Fatal("different large integers share the same key")
	}
}
