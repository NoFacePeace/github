package tool

import (
	"context"
	"encoding/json"
)

type Tool interface {
	Definition() Definition
	// Execute 的实现负责响应 ctx 的取消和超时，并返回对应错误。
	Execute(ctx context.Context, args json.RawMessage) (Result, error)
}
