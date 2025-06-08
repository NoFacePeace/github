package runnable

import "context"

type Runnable interface {
	Start(context.Context) error
}
