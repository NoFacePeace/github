package runner

import (
	"context"
	"errors"
	"time"

	"infra-agent-cloud/executor"
	"infra-agent-cloud/protocol"
)

// ErrExecutionLeaseRejected 表示 Server 已明确拒绝当前执行的续约。
var ErrExecutionLeaseRejected = errors.New("execution lease rejected")

// Client 封装 Runner 对 Server 的 HTTP 调用。
type Client interface {
	// RenewExecution 返回确认的过期时间，明确拒绝时返回 ErrExecutionLeaseRejected。
	RenewExecution(ctx context.Context, job *protocol.Job) (time.Time, error)
	// ClaimSecrets 领取本轮凭证，无凭证时返回 nil, nil。
	ClaimSecrets(ctx context.Context, job *protocol.Job) (*executor.Secrets, error)
}
