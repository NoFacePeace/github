package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"infra-knowledge/api"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, *addr); err != nil {
		log.Printf("HTTP server failed: %v", err)
		os.Exit(1)
	}
}

// run 启动 HTTP 服务，收到取消信号后等待存量请求完成。
func run(ctx context.Context, addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	server := &http.Server{
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Serve(listener)
	}()
	log.Printf("HTTP server listening on %s", listener.Addr())

	select {
	case err := <-serverErr:
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		// 退出信号已取消主上下文，关闭流程使用独立且有时限的上下文。
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			// 超时后强制关闭连接，避免进程遗留正在处理的请求。
			_ = server.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		if err := <-serverErr; !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		log.Print("HTTP server stopped")
		return nil
	}
}
