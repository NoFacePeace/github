package executor

import "context"

// ExecutionSpec 描述本次引擎执行的规格，具体配置字段后续补充。
type ExecutionSpec struct{}

// ExecutionSecrets 保存通过 stdin 管道传入的敏感执行配置，具体字段后续补充。
type ExecutionSecrets struct{}

// Result 是引擎执行结果，后续补充引擎会话 ID、退出码等字段。
type Result struct{}

// Event 是执行期间的引擎输出，具体协议字段后续补充。
type Event struct{}

// Executor 封装具体引擎调用，不依赖 Server 或 Runner 的业务模型。
type Executor interface {
	// Run 按执行规格运行引擎并通过 emit 输出事件；取消后须等待进程退出再返回。
	Run(ctx context.Context, spec *ExecutionSpec, emit func(Event) error) (*Result, error)
}
