package agentloop

import "context"

// Recorder 定义执行循环的过程记录能力。
type Recorder interface {
	Record(ctx context.Context, entry Entry) error
}
