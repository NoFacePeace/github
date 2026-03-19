package core

import (
	"context"

	umov1 "nofacepeace.github.io/controller/api/v1"
)

type ServiceManager struct {
}

func (s *ServiceManager) CheckService(ctx context.Context, cls *umov1.Middleware) error {
	return nil
}
