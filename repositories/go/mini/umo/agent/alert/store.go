package alert

import "context"

type Store interface {
	GetAlert(ctx context.Context, id string) (*Alert, error)
	GetBinding(ctx context.Context, id string) (*Binding, error)
	CreateAnalysis(ctx context.Context, analysis *Analysis) error
	UpdateAnalysis(ctx context.Context, args ...any) error
}
