package ossignal

import (
	"context"
	"os"
	"os/signal"
)

func WaitOSInterrupt() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer func() {
		signal.Stop(ch)
		close(ch)
	}()
	<-ch
}

func WaitOSInterruptWithContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx
}
