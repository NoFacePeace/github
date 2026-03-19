package core

import (
	"context"

	"nofacepeace.github.io/controller/pkg/extensions/event"
	"nofacepeace.github.io/controller/pkg/model"
)

type EventManager struct {
}

func (e *EventManager) Dispatch(ctx context.Context, cls string, typ model.EventType, nodes []*event.Node) {
}
