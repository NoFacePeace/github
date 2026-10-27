package provider

import "context"

type Provider interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
