package executor

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// Engine 提供公共执行流程所需的引擎专属信息。
type Engine interface {
	// WriteConfig 写入本轮所需的引擎配置和认证文件。
	WriteConfig(spec *Spec, secrets *Secrets) error
	// BuildCommand 生成启动参数，供公共执行流程启动进程。
	BuildCommand(spec *Spec, secrets *Secrets) (EngineCommand, error)
	// ParseOutput 收集一行引擎标准输出中的结果信息。
	ParseOutput(line string)
	// ParseResult 在输出读完后填充解析结果，并报告协议或执行错误。
	ParseResult(result *Result) error
}

// EngineCommand 描述引擎的启动命令，Env 保存需要覆盖的环境变量。
type EngineCommand struct {
	Executable string
	Args       []string
	Dir        string
	Env        []string
}

// Run 是公共执行入口，commands 为可选的执行中输入，普通模式传 nil。
func Run(ctx context.Context, spec *Spec, secrets *Secrets, commands <-chan SteerCommand, emit func(Event) error) (*Result, error) {
	runCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var mu sync.Mutex
	var seq int64
	var emitErr error
	publish := func(kind EventType, payload map[string]any) error {
		mu.Lock()
		defer mu.Unlock()
		if emitErr != nil {
			return emitErr
		}
		seq++
		if err := emit(Event{
			SchemaVersion: spec.SchemaVersion,
			ExecutionID:   spec.ExecutionID,
			Seq:           seq,
			Timestamp:     time.Now().UTC(),
			Type:          kind,
			Payload:       payload,
		}); err != nil {
			emitErr = fmt.Errorf("emit %s event: %w", kind, err)
			cancel(emitErr)
		}
		return emitErr
	}

	result, runErr := runEngine(runCtx, spec, secrets, publish)
	if result == nil {
		result = &Result{}
	}
	if emitErr != nil {
		return result, emitErr
	}
	kind := EventTypeCompleted
	if errors.Is(ctx.Err(), context.Canceled) {
		kind = EventTypeStopped
		runErr = context.Cause(ctx)
	} else if runErr != nil {
		kind = EventTypeFailed
		result.Error = runErr.Error()
	}
	// 取消后仍发送终态。若输出回调已失败，则由前面的检查直接返回。
	if err := publish(kind, map[string]any{"result": result}); err != nil {
		return result, errors.Join(runErr, err)
	}
	return result, runErr
}

func runEngine(ctx context.Context, spec *Spec, secrets *Secrets, publish func(EventType, map[string]any) error) (*Result, error) {
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	var engine Engine
	switch spec.Engine {
	case "codex":
		engine = NewCodex()
	default:
		return nil, fmt.Errorf("unsupported engine: %q", spec.Engine)
	}

	command, err := engine.BuildCommand(spec, secrets)
	if err != nil {
		return nil, fmt.Errorf("build %s command: %w", spec.Engine, err)
	}
	path, err := exec.LookPath(command.Executable)
	if err != nil {
		return nil, fmt.Errorf("cannot find %s executable: %w", command.Executable, err)
	}
	command.Executable = path
	if err := engine.WriteConfig(spec, secrets); err != nil {
		return nil, fmt.Errorf("prepare %s config: %w", spec.Engine, err)
	}

	return runEngineProcess(ctx, engine, command, publish)
}

// Executor 封装具体引擎调用，不依赖 Server 或 Runner 的业务模型。
type Executor interface {
	// Run 按执行规格运行引擎并通过 emit 输出事件；取消后须等待进程退出再返回。
	Run(ctx context.Context, spec *Spec, emit func(Event) error) (*Result, error)
}
