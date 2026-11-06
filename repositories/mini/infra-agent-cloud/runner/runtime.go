package runner

import (
	"context"
	"errors"
	"infra-agent-cloud/executor"
	"time"
)

// Runtime 承载任务的执行环境，负责工作区准备、执行及资源清理。
type Runtime interface {
	// Run 执行本轮任务，ctx 取消时须停止执行并完成资源清理后返回。
	Run(ctx context.Context, job *Job, emit func(executor.Event) error) (*executor.Result, error)
}

type runtime struct {
	executor executor.Executor
}

func NewRuntime(engine executor.Executor) Runtime {
	return &runtime{executor: engine}
}

// Run 编排本轮执行流程：
//  1. 准备或恢复任务工作区。
//  2. 创建 cgroup 并配置资源限额，准备隔离环境（待接入）。
//  3. 调用 Executor，在上述环境内启动引擎并回传输出；Context 取消时停止并等待进程退出。
//  4. 清理残留进程、cgroup 和隔离资源（待接入），再清理本轮工作区临时资源，保留任务文件和会话。
//
// cgroup 归属须在引擎及其子进程运行前生效，不能等进程启动后再补加限制。
func (r *runtime) Run(ctx context.Context, job *Job, emit func(executor.Event) error) (result *executor.Result, err error) {
	request, err := r.prepareWorkspace(ctx, job)
	if err != nil {
		return nil, err
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		err = errors.Join(err, r.cleanupWorkspace(cleanupCtx, job))
	}()
	return r.executor.Run(ctx, request, emit)
}

func (r *runtime) prepareWorkspace(ctx context.Context, job *Job) (*executor.Request, error) {
	// TODO: 准备或恢复任务目录，将 Job 映射为引擎输入；部分失败时回收已创建的临时资源。
	return nil, errors.New("workspace preparation is not implemented")
}

func (r *runtime) cleanupWorkspace(ctx context.Context, job *Job) error {
	// TODO: 清理本轮临时资源，保留任务工作区及继续执行所需的会话文件。
	return errors.New("workspace cleanup is not implemented")
}
