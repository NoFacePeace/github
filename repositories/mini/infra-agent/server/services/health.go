// Package services 提供应用业务能力。
package services

import "context"

// HealthService 提供健康检查业务能力。
type HealthService struct{}

// NewHealthService 创建健康检查服务。
func NewHealthService() *HealthService {
	return &HealthService{}
}

// Ping 返回服务连通性检查结果。
func (s *HealthService) Ping(_ context.Context) string {
	return "pong"
}
