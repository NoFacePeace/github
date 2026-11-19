package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"umo/controller/healthcheck"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// 具体检查器与配置在对应功能实现时接入。
	manager, err := healthcheck.NewManager(nil, healthcheck.Config{})
	if err != nil {
		stop()
		log.Printf("create healthcheck manager failed: %v", err)
		os.Exit(1)
	}
	managerDone := make(chan struct{})
	go func() {
		defer close(managerDone)
		manager.Run(ctx)
	}()

	// 等待管理器完成退出，确保运行资源已清理。
	<-managerDone
	stop()
}
