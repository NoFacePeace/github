package healthcheck

import (
	"context"
	"errors"
)

// Checker 定义具体健康检查器的调用约定，只检测并返回异常，不修改资源。
type Checker interface {
	GetName() string
	// Check 返回 Pod 状态及检查错误，两者可以同时存在。
	// 键使用 namespace/name，每个 Pod 返回一个异常状态，正常 Pod 不写入结果。
	Check(ctx context.Context) (map[string]Status, error)
}

// BuiltinHealthChecker 从 Kubernetes 对象读取健康状态。
type BuiltinHealthChecker struct{}

func (BuiltinHealthChecker) GetName() string { return "builtin" }

func (BuiltinHealthChecker) Check(context.Context) (map[string]Status, error) {
	return nil, errors.New("builtin health checker is not implemented")
}

// WebhookHealthChecker 从外部 HTTP 接口读取健康状态。
type WebhookHealthChecker struct{}

func (WebhookHealthChecker) GetName() string { return "webhook" }

func (WebhookHealthChecker) Check(context.Context) (map[string]Status, error) {
	return nil, errors.New("webhook health checker is not implemented")
}
