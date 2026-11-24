package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"infra-agent-cloud/executor"
)

const (
	// commandRun 执行任务。
	commandRun = "run"
	// commandPreflight 检查执行环境。
	commandPreflight = "preflight"
	// commandVersion 输出程序版本。
	commandVersion = "version"
)

// main 是独立 Executor 程序的入口。
func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if code != 0 {
		os.Exit(code)
	}
}

// run 在返回前执行输入通道和信号监听的清理，再由 main 设置退出码。
func run() (int, error) {
	// TODO: 按以下顺序接入程序入口流程：
	//  1. 解析 -command 参数：run 执行任务，preflight 检查执行环境，version 输出程序版本。
	command := flag.String("command", "", "Command to execute: run, preflight, version")
	manifest := flag.String("manifest", "", "Path to the execution manifest JSON file")
	flag.Parse()
	switch *command {
	case commandRun:
	case commandPreflight:
		return 0, nil
	case commandVersion:
		return 0, nil
	default:
		return 0, nil
	}

	// 2. run 读取 --manifest 指定的 JSON 执行清单，校验协议版本和执行参数。
	data, err := os.ReadFile(*manifest)
	if err != nil {
		return 1, fmt.Errorf("failed to read execution manifest: %w", err)
	}
	var spec executor.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return 1, fmt.Errorf("failed to parse execution manifest: %w", err)
	}
	// TODO: 补齐执行规格字段后校验协议版本和执行参数。

	// 3. 在读取 stdin 前创建执行 Context，收到 SIGINT 或 SIGTERM 时取消读取和执行。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. 按模式读取敏感配置，steer 模式首帧返回后继续通过命令流读取后续输入。
	input, err := openInput(os.Stdin)
	if err != nil {
		return 1, err
	}
	defer input.Close()
	executionCtx := ctx
	var secrets *executor.Secrets
	var commands <-chan executor.SteerCommand
	switch spec.SteerPromptMode {
	case "", executor.SteerPromptModeLegacy:
		secrets, err = executor.ReadSecrets(ctx, input)
		if err != nil {
			return 1, fmt.Errorf("failed to read execution secrets: %w", err)
		}
	case executor.SteerPromptModeSteer:
		control, err := executor.OpenControlInput(ctx, input, spec.ExecutionID)
		if err != nil {
			return 1, fmt.Errorf("failed to open control input: %w", err)
		}
		defer control.Close()
		secrets = control.Secrets
		commands = control.Commands
		executionCtx = control.Context
	default:
		return 1, fmt.Errorf("unsupported steer prompt mode")
	}
	// TODO: 校验普通模式下敏感配置的协议版本和执行归属。

	// 5. 准备事件输出回调：将统一事件编码为 JSONL 写入 stdout，诊断信息写入 stderr。
	encoder := json.NewEncoder(os.Stdout)
	emit := func(event executor.Event) error {
		return encoder.Encode(event)
	}
	// 6. 调用执行入口，传入执行参数、Context、敏感配置、事件回调和可选控制消息通道。
	result, err := executor.Run(executionCtx, &spec, secrets, commands, emit)
	// 7. 返回退出码，实际退出前会关闭控制输入并释放信号监听。
	if err != nil {
		code := 1
		if ctx.Err() == context.Canceled {
			code = 130
		} else if result.ExitCode != nil && *result.ExitCode > 0 {
			code = *result.ExitCode
		}
		return code, fmt.Errorf("execution failed: %w", err)
	}
	return 0, nil
}

// openInput 将继承的阻塞输入描述符复制为可轮询文件，使 Close 能解除正在等待的 Read。
func openInput(source *os.File) (*os.File, error) {
	fd, err := syscall.Dup(int(source.Fd()))
	if err == nil {
		syscall.CloseOnExec(fd)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate execution input: %w", err)
	}
	if err := syscall.SetNonblock(fd, true); err != nil {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("failed to prepare execution input: %w", err)
	}
	return os.NewFile(uintptr(fd), "execution-input"), nil
}
