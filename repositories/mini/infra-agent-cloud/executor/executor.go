package executor

import (
	"context"
	"errors"
)

// Run 是公共执行入口，commands 为可选的执行中输入，普通模式传 nil。
func Run(ctx context.Context, spec *Spec, secrets *Secrets, commands <-chan SteerCommand, emit func(Event) error) (*Result, error) {
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	// TODO: 选择引擎、准备配置、启动进程、消费控制命令，并通过 emit 回传事件和收敛终态。
	return nil, errors.New("executor execution is not implemented")
}

// Executor 封装具体引擎调用，不依赖 Server 或 Runner 的业务模型。
type Executor interface {
	// Run 按执行规格运行引擎并通过 emit 输出事件；取消后须等待进程退出再返回。
	Run(ctx context.Context, spec *Spec, emit func(Event) error) (*Result, error)
}
