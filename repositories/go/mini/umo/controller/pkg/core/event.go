package core

import (
	"context"
	"errors"

	"github.com/NoFacePeace/github/repositories/go/utils/goroutine"
	"nofacepeace.github.io/controller/pkg/config"
	"nofacepeace.github.io/controller/pkg/extensions/event"
	"nofacepeace.github.io/controller/pkg/model"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type EventManager struct {
	listeners []event.Listener
	chanMap   map[string]chan event.Event
}

func NewEventManager(ls []event.Listener) *EventManager {
	m := map[string]chan event.Event{}
	for _, listener := range ls {
		m[listener.GetName()] = make(chan event.Event, 100)
	}
	mg := &EventManager{
		listeners: ls,
		chanMap:   m,
	}
	return mg
}

func (m *EventManager) Start() {
	for _, listener := range m.listeners {
		l := listener
		ch := m.chanMap[l.GetName()]
		goroutine.Recover(func() {
			for event := range ch {
				l.OnEvent(event)
			}
		})
	}
}

func (m *EventManager) Dispatch(ctx context.Context, cls string, typ model.EventType, nodes []*event.Node) {
	logger := logf.FromContext(ctx)
	if config.InDryRunMode(cls) {
		return
	}
	e := event.Event{}
	for name, ch := range m.chanMap {
		select {
		case ch <- e:
		default:
			logger.Error(errors.New("channel full"), "channel full", "type", typ, "listener", name, "cluster", cls)
		}
	}
}
