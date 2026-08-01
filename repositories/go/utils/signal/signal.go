// Package signal 提供旧版的进程级退出信号处理。
//
// Deprecated: 本包通过 sync.Once 创建唯一的全局 context，调用方不能配置父
// context 或监听的信号，也不能取消信号注册；测试之间也无法重置包状态。收到
// 第二次信号时会调用 os.Exit(1)，因此会跳过 defer 清理。请在程序入口直接使用
// os/signal.NotifyContext；可参考 demo/signal-context 中支持优雅退出及第二次
// Ctrl-C 强制终止的示例。
package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var once sync.Once
var ctx context.Context

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func SetupSignalHandler() context.Context {
	once.Do(func() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(context.Background())
		c := make(chan os.Signal, 2)
		signal.Notify(c, shutdownSignals...)
		go func() {
			<-c
			cancel()
			<-c
			os.Exit(1)
		}()
	})
	return ctx
}
