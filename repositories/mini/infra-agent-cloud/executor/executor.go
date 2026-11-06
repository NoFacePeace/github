package executor

import "context"

// Request 是引擎执行输入，具体配置字段后续补充。
type Request struct{}

// Result 是引擎执行结果，后续补充引擎会话 ID、退出码等字段。
type Result struct{}

// Event 是执行期间的引擎输出，具体协议字段后续补充。
type Event struct{}

// Executor 封装具体引擎调用，不依赖 Server 或 Runner 的业务模型。
type Executor interface {
	// Run 执行请求并通过 emit 输出事件；取消后须等待进程退出再返回。
	Run(ctx context.Context, request *Request, emit func(Event) error) (*Result, error)
}
