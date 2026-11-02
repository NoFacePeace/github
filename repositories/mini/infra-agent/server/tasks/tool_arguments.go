package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// validateToolArguments 编译工具的 JSON Schema，并校验调用参数。
func validateToolArguments(input, schema json.RawMessage) error {
	if !json.Valid(input) {
		return fmt.Errorf("arguments must be valid JSON")
	}
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(input))
	if err != nil {
		return fmt.Errorf("decode arguments: %w", err)
	}
	definition, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return fmt.Errorf("decode tool schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	// Schema 随工具提供，只解析内存资源及内部引用，不加载外部文件或网络资源。
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	const schemaURL = "urn:infra-agent:tool-arguments"
	if err := compiler.AddResource(schemaURL, definition); err != nil {
		return fmt.Errorf("load tool schema: %w", err)
	}
	compiled, err := compiler.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("compile tool schema: %w", err)
	}
	if err := compiled.Validate(value); err != nil {
		return fmt.Errorf("invalid tool arguments: %w", err)
	}
	return nil
}
