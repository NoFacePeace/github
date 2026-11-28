package runner

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

func (r *Runner) startExecutionRenewal(ctx context.Context, job *Job, stopExecution context.CancelFunc) (func(), error) {
	expiresAt, err := r.renewExecution(ctx, job, job.LeaseExpiresAt)
	if err != nil {
		return nil, err
	}

	renewCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		retrying := false
		for {
			remaining := time.Until(expiresAt)
			if remaining <= 0 {
				stopExecution()
				return
			}
			delay := min(10*time.Second, remaining/2)
			if retrying {
				delay = min(time.Second, remaining)
			}
			timer := time.NewTimer(delay)
			select {
			case <-renewCtx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			next, err := r.renewExecution(renewCtx, job, expiresAt)
			if renewCtx.Err() != nil {
				return
			}
			if errors.Is(err, ErrExecutionLeaseRejected) || !time.Now().Before(expiresAt) {
				stopExecution()
				return
			}
			if err != nil {
				slog.Warn("Failed to renew execution lease", "executionID", job.MessageID)
				retrying = true
				continue
			}
			expiresAt = next
			retrying = false
		}
	}()
	return func() {
		cancel()
		<-done
	}, nil
}

func (r *Runner) renewExecution(ctx context.Context, job *Job, expiresAt time.Time) (time.Time, error) {
	if !time.Now().Before(expiresAt) {
		return time.Time{}, ErrExecutionLeaseRejected
	}
	renewCtx, cancel := context.WithDeadline(ctx, expiresAt)
	defer cancel()
	next, err := r.client.RenewExecution(renewCtx, job)
	if err != nil {
		return time.Time{}, err
	}
	if err := renewCtx.Err(); err != nil {
		return time.Time{}, err
	}
	if !next.After(time.Now()) {
		return time.Time{}, ErrExecutionLeaseRejected
	}
	return next, nil
}
