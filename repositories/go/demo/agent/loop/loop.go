package loop

import "context"

type Loop interface {
	Run(ctx context.Context, req Request, onEvent func(Event)) error
}
