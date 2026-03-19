package core

import (
	"context"

	umov1 "nofacepeace.github.io/controller/api/v1"
)

type StatusManager struct {
}

func (s *StatusManager) UpdateStatus(ctx context.Context, cls *umov1.Middleware) error {
	return nil
}

func (s *StatusManager) updateClusterPhase(ctx context.Context, cls *umov1.Middleware, phase umov1.MiddlewarePhase) error {
	return nil
}

func (s *StatusManager) updateClusterChecksum(ctx context.Context, cls *umov1.Middleware) error {
	return nil
}
