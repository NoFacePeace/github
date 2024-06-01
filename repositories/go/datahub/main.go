package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/NoFacePeace/github/repositories/go/datahub/stock/tencent"
	"github.com/NoFacePeace/github/repositories/go/utils/config"
	"github.com/NoFacePeace/github/repositories/go/utils/log"
	"github.com/robfig/cron/v3"
)

func main() {
	log.Init()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// config
	var cfg Config
	if err := config.ReadYamlFile("config.yaml", &cfg); err != nil {
		slog.ErrorContext(ctx, err.Error())
		return
	}
	slog.InfoContext(ctx, "config loaded")
	address := "http://localhost:8428/api/v1/import"
	vm := tencent.NewVictoriaMetrics(address)
	// tencent
	tc := tencent.New(vm)
	go func() {
		tc.History()
		// tc.Daily()
	}()
	c := cron.New()
	c.Start()
	slog.InfoContext(ctx, "cron started")
	slog.InfoContext(ctx, "process started")
	// process stop
	<-ctx.Done()
	stop()
	// cron stop
	c.Stop()
	slog.InfoContext(ctx, "process stopped")
}

type Config struct {
	DB struct {
		DSN  string
		Type string
	}
}
