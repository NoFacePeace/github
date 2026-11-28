package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"infra-agent-cloud/protocol"
	"infra-agent-cloud/runner"
)

// main 是 Runner 程序的入口。
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// TODO: 读取 Runner 配置并注入 Client、Store 和 Runtime 的具体实现。
	node := runner.NewRunner(protocol.Heartbeat{}, nil, nil, nil)
	err := node.Run(ctx)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
