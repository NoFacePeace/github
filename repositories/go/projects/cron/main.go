package main

import (
	"log/slog"
	"os"

	"github.com/NoFacePeace/github/repositories/go/projects/cron/channel"
	"github.com/NoFacePeace/github/repositories/go/projects/cron/task"
	"github.com/NoFacePeace/github/repositories/go/utils/config"
	"github.com/NoFacePeace/github/repositories/go/utils/signal"
)

func main() {
	ctx := signal.SetupSignalHandler()
	mgr := task.NewManager()
	cfg := &Config{}
	if err := config.ReadYamlFile("config.yaml", cfg); err != nil {
		slog.Error("config read yaml file error", "error", err)
	}
	for _, v := range cfg.Tasks {
		chs := []channel.Channel{}
		for _, c := range v.Channels {
			ch, err := channel.New(c.Name, c.Config)
			if err != nil {
				slog.Error("channel new error", "error", err)
				os.Exit(1)
			}
			chs = append(chs, ch)
		}
		t := task.New(v.Name, v.Spec, chs, task.GetPlatesDroppedBy20Percent)
		mgr.Add(t)
	}
	if err := mgr.Start(ctx); err != nil {
		slog.Error("task manager start error", "error", err)
		os.Exit(1)
	}
}

type Config struct {
	Tasks []task.Config
}
