package tasks

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateToolArguments(t *testing.T) {
	for _, tt := range []struct {
		name, input, schema, errorPath string
	}{
		{name: "required string", input: `{"path":"a.go"}`, schema: `{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`},
		{name: "missing required", input: `{}`, schema: `{"type":"object","required":["path"]}`, errorPath: "$.path"},
		{name: "wrong type", input: `{"path":12}`, schema: `{"properties":{"path":{"type":"string"}}}`, errorPath: "$.path"},
		{name: "nested field", input: `{"options":{"limit":"a"}}`, schema: `{"properties":{"options":{"properties":{"limit":{"type":"integer"}}}}}`, errorPath: "$.options.limit"},
		{name: "extra field", input: `{"path":"a","extra":1}`, schema: `{"properties":{"path":{"type":"string"}},"additionalProperties":false}`, errorPath: "$.extra"},
		{name: "default allows extra", input: `{"path":"a","extra":1}`, schema: `{"properties":{"path":{"type":"string"}}}`},
		{name: "allow extra", input: `{"path":"a","extra":1}`, schema: `{"properties":{"path":{"type":"string"}},"additionalProperties":true}`},
		{name: "forbid all extras", input: `{"extra":1}`, schema: `{"additionalProperties":false}`, errorPath: "$.extra"},
		{name: "typed extras", input: `{"extra":1}`, schema: `{"additionalProperties":{"type":"integer"}}`},
		{name: "invalid typed extra", input: `{"extra":"a"}`, schema: `{"additionalProperties":{"type":"integer"}}`, errorPath: "$.extra"},
		{name: "array element", input: `{"paths":["a",1]}`, schema: `{"properties":{"paths":{"type":"array","items":{"type":"string"}}}}`, errorPath: "$.paths[1]"},
		{name: "array minimum", input: `[]`, schema: `{"type":"array","minItems":1}`, errorPath: "$"},
		{name: "array maximum", input: `[1,2]`, schema: `{"maxItems":1}`, errorPath: "$"},
		{name: "unicode length", input: `"你好"`, schema: `{"type":"string","minLength":2,"maxLength":2}`},
		{name: "short string", input: `"a"`, schema: `{"minLength":2}`, errorPath: "$"},
		{name: "long string", input: `"abc"`, schema: `{"maxLength":2}`, errorPath: "$"},
		{name: "pattern match", input: `"abc"`, schema: `{"pattern":"^[a-z]+$"}`},
		{name: "pattern mismatch", input: `"123"`, schema: `{"pattern":"^[a-z]+$"}`, errorPath: "$"},
		{name: "invalid pattern", input: `"a"`, schema: `{"pattern":"["}`, errorPath: "schema"},
		{name: "numeric limits inclusive", input: `10`, schema: `{"type":"number","minimum":10,"maximum":10}`},
		{name: "below minimum", input: `9`, schema: `{"minimum":10}`, errorPath: "$"},
		{name: "above maximum", input: `11`, schema: `{"maximum":10}`, errorPath: "$"},
		{name: "large integer", input: `9007199254740993`, schema: `{"type":"integer","minimum":9007199254740993,"maximum":9007199254740993}`},
		{name: "large integer lower bound", input: `9007199254740992`, schema: `{"minimum":9007199254740993}`, errorPath: "$"},
		{name: "decimal integer", input: `2.0`, schema: `{"type":"integer"}`},
		{name: "fraction not integer", input: `2.5`, schema: `{"type":"integer"}`, errorPath: "$"},
		{name: "null union", input: `null`, schema: `{"type":["string","null"]}`},
		{name: "boolean", input: `true`, schema: `{"type":"boolean"}`},
		{name: "enum match", input: `"read"`, schema: `{"enum":["read","write"]}`},
		{name: "enum mismatch", input: `"delete"`, schema: `{"enum":["read","write"]}`, errorPath: "$"},
		{name: "numeric enum", input: `1.0`, schema: `{"enum":[1]}`},
		{name: "large enum mismatch", input: `9007199254740992`, schema: `{"enum":[9007199254740993]}`, errorPath: "$"},
		{name: "nested enum", input: `{"a":[1.0]}`, schema: `{"enum":[{"a":[1]}]}`},
		{name: "empty enum", input: `1`, schema: `{"enum":[]}`, errorPath: "$"},
		{name: "allOf", input: `5`, schema: `{"allOf":[{"type":"integer"},{"minimum":1},{"maximum":10}]}`},
		{name: "allOf mismatch", input: `11`, schema: `{"allOf":[{"minimum":1},{"maximum":10}]}`, errorPath: "$"},
		{name: "anyOf", input: `5`, schema: `{"anyOf":[{"type":"integer"},{"type":"string"}]}`},
		{name: "anyOf mismatch", input: `false`, schema: `{"anyOf":[{"type":"integer"},{"type":"string"}]}`, errorPath: "$"},
		{name: "oneOf", input: `"a"`, schema: `{"oneOf":[{"type":"number"},{"type":"string"}]}`},
		{name: "oneOf overlap", input: `2`, schema: `{"oneOf":[{"type":"number"},{"type":"integer"}]}`, errorPath: "$"},
		{name: "invalid JSON", input: `{`, schema: `{}`, errorPath: "valid JSON"},
		{name: "trailing JSON", input: `{} {}`, schema: `{}`, errorPath: "valid JSON"},
		{name: "invalid schema", input: `{}`, schema: `{`, errorPath: "schema"},
		{name: "null schema", input: `{}`, schema: `null`, errorPath: "schema"},
		{name: "invalid type declaration", input: `1`, schema: `{"type":12}`, errorPath: "schema"},
		{name: "unknown type", input: `1`, schema: `{"type":"unknown"}`, errorPath: "schema"},
		{name: "internal reference", input: `{"path":"a.go"}`, schema: `{"$defs":{"path":{"type":"string"}},"properties":{"path":{"$ref":"#/$defs/path"}}}`},
		{name: "invalid internal reference value", input: `{"path":123}`, schema: `{"$defs":{"path":{"type":"string"}},"properties":{"path":{"$ref":"#/$defs/path"}}}`, errorPath: "$.path"},
		{name: "external reference not loaded", input: `{}`, schema: `{"$ref":"file:///tmp/tool-schema.json"}`, errorPath: "schema"},
		{name: "boolean true schema", input: `{}`, schema: `true`},
		{name: "boolean false schema", input: `{}`, schema: `false`, errorPath: "$"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToolArguments(json.RawMessage(tt.input), json.RawMessage(tt.schema))
			if tt.errorPath == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				// 库使用 JSON Pointer 表示路径，必填和额外属性错误可能附在父对象上。
				text := strings.TrimPrefix(tt.errorPath, "$.")
				text = strings.ReplaceAll(text, ".", "/")
				text = strings.ReplaceAll(text, "[", "/")
				text = strings.ReplaceAll(text, "]", "")
				if text == "$" {
					text = "invalid tool arguments"
				}
				if err == nil || !strings.Contains(err.Error(), text) {
					t.Fatalf("got %v, want error containing %q", err, text)
				}
			}
		})
	}
}
