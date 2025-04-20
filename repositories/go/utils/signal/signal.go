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
