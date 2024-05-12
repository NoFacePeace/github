package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/NoFacePeace/github/repositories/go/datahub/stock/tencent"
	"github.com/NoFacePeace/github/repositories/go/util/config"
	"github.com/NoFacePeace/github/repositories/go/util/log"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
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
	// db
	db, err := gorm.Open(clickhouse.Open(cfg.ClickHouse.DSN), &gorm.Config{})
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
	}
	slog.InfoContext(ctx, "db connected")
	// tencent
	tc := tencent.New(db)
	go func() {
		tc.History()
	}()
	c := cron.New()
	c.AddFunc("0 18 * * *", tc.Daily)
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
	ClickHouse struct {
		DSN string
	}
}
