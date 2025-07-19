package task

import (
	"context"
	"log/slog"

	"github.com/NoFacePeace/github/repositories/go/projects/cron/channel"
	"github.com/robfig/cron/v3"
)

type Manager interface {
	Start(context.Context) error
	Add(*Task)
}

type manager struct {
	tasks []*Task
}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) Start(ctx context.Context) error {
	c := cron.New()
	for _, task := range m.tasks {
		fn := func() {
			slog.Info("task start")
			defer slog.Info("task end")
			ret := task.Start()
			if task.lastResult != nil && task.lastResult.Message == ret.Message {
				return
			}
			task.lastResult = &ret
			msg := &channel.Message{}
			msg.Title = task.name
			msg.Content = ret.String()
			for _, channel := range task.Channels() {
				if err := channel.Send(msg); err != nil {
					slog.Error("channel send error", "error", err)
				}
			}
		}
		fn()
		c.AddFunc(task.Spec(), fn)
	}
	c.Start()
	<-ctx.Done()
	c.Stop()
	return nil
}

func (m *manager) Add(t *Task) {
	m.tasks = append(m.tasks, t)
}
