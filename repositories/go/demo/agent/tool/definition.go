package tool

import "encoding/json"

type Definition struct {
	Name        string
	Description string
	InputSchema json.RawMessage // 参数的 JSON Schema
}
