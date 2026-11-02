package tasks

import (
	"context"
	"log/slog"
	"time"
)

// lease 表示本轮执行持有的租约凭证。
type lease struct {
	Owner     string
	Epoch     int64
	ExpiresAt time.Time
}

// renewLease 定期续租，失败时取消执行；调用方负责启动和等待该方法退出。
func (s *service) renewLease(ctx context.Context, runID string, execution *execution, ttl time.Duration) {
	interval := ttl / 3
	if interval <= 0 {
		interval = time.Nanosecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
			renewCtx, cancel := context.WithTimeout(ctx, interval)
			renewed, err := s.store.RenewLease(renewCtx, runID, execution.lease.Owner, execution.lease.Epoch, ttl)
			cancel()
			if ctx.Err() != nil {
				return
			}
			if err != nil || !renewed {
				execution.leaseLost.Store(true)
				slog.ErrorContext(ctx, "renew run lease failed", "run_id", runID, "renewed", renewed, "error", err)
				execution.cancel()
				return
			}
		}
	}
}
