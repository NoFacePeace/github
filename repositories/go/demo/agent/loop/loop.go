package loop

import (
	"context"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/event"
)

type Loop interface {
	Run(ctx context.Context, req Request, onEvent func(event.Event)) error
}
