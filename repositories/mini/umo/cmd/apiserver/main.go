package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"umo/apiserver"
	"umo/apiserver/watch"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	var store watch.Store
	broadcaster := watch.NewEventBroadcaster()
	var wg sync.WaitGroup
	defer func() {
		stop()
		wg.Wait()
	}()
	if store != nil {
		wg.Go(func() {
			watch.RunPoller(ctx, store, broadcaster)
		})
	}

	server := &http.Server{
		Addr:        ":8080",
		Handler:     apiserver.NewRouter(store, broadcaster),
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve API: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shut down API server: %v", err)
		}
	}
}
