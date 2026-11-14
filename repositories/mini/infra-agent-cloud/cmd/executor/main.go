package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

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
	// TODO: 按以下顺序接入程序入口流程：
	//  1. 解析 -command 参数：run 执行任务，preflight 检查执行环境，version 输出程序版本。
	command := flag.String("command", "", "Command to execute: run, preflight, version")
	manifest := flag.String("manifest", "", "Path to the execution manifest JSON file")
	flag.Parse()
	switch *command {
	case commandRun:
	case commandPreflight:
		return
	case commandVersion:
		return
	default:
		return
	}

	// 2. run 读取 --manifest 指定的 JSON 执行清单，校验协议版本和执行参数。
	data, err := os.ReadFile(*manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read execution manifest: %v\n", err)
		os.Exit(1)
	}
	var spec executor.ExecutionSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse execution manifest: %v\n", err)
		os.Exit(1)
	}
	// TODO: 补齐执行规格字段后校验协议版本和执行参数。

	// 3. 从 stdin 读取敏感执行配置；后续支持 steer 时，同时建立控制消息读取通道。
	secretData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read execution secrets")
		os.Exit(1)
	}
	var secrets executor.ExecutionSecrets
	if err := json.Unmarshal(secretData, &secrets); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse execution secrets")
		os.Exit(1)
	}
	// TODO: 补齐敏感配置字段后校验协议版本和执行归属。

	// 4. 创建执行 Context，收到 SIGINT 或 SIGTERM 时取消执行。
	// 5. 准备事件输出回调：将统一事件编码为 JSONL 写入 stdout，诊断信息写入 stderr。
	// 6. 调用执行入口，传入执行参数、Context、敏感配置、事件回调和可选控制消息通道。
	// 7. 执行结束后关闭控制消息读取通道，释放信号监听，并以执行结果设置进程退出码。
}
