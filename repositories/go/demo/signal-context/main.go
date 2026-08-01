// signal-context 演示如何使用 signal.NotifyContext 优雅退出。
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // main 返回时取消信号注册，释放相关资源。

	go func() {
		<-ctx.Done() // 第一次 Ctrl-C 或 SIGTERM：通知服务优雅退出。
		stop()       // 恢复默认信号行为；第二次 Ctrl-C 会立刻终止进程。
	}()

	log.Println("服务已启动；按 Ctrl-C 优雅退出")
	if err := run(ctx); err != nil {
		log.Printf("服务因错误停止：%v", err)
	}
}

// run 代表一个阻塞运行的服务循环。它负责等待 ctx.Done()，因此 main
// 无需额外执行 <-ctx.Done()。第一次信号取消 context 后，调用方会恢复
// 默认信号行为，使第二次 Ctrl-C 可以强制终止进程。
func run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("收到退出信号：%v", ctx.Err())
			return nil
		case <-ticker.C:
			log.Println("服务运行中")
		}
	}
}
