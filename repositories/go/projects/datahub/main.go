package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/NoFacePeace/github/repositories/go/projects/datahub/stock/tencent"
	"github.com/NoFacePeace/github/repositories/go/utils/log"

	"github.com/robfig/cron/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	log.Init()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// config
	// var cfg Config
	// if err := config.ReadYamlFile("config.yaml", &cfg); err != nil {
	// 	slog.ErrorContext(ctx, err.Error())
	// 	return
	// }
	slog.InfoContext(ctx, "config loaded")
	db, err := gorm.Open(mysql.Open("root:@tcp(127.0.0.1:3306)/db_stock?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return
	}
	mysql := tencent.NewMySQL(db)
	// tencent
	tc := tencent.New(mysql)
	go func() {
		tc.History()
		tc.Daily()
	}()
	c := cron.New()
	c.AddFunc("0 16 * * *", tc.Daily)
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
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
}
