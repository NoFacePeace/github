package tasks

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

type renewalStore struct {
	stubTaskStore
	renew func(context.Context, string, string, int64, time.Duration) (bool, error)
}

func (s *renewalStore) RenewLease(ctx context.Context, id, owner string, epoch int64, ttl time.Duration) (bool, error) {
	return s.renew(ctx, id, owner, epoch, ttl)
}

func TestRenewLeaseSuccessUntilCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		s := &service{store: &renewalStore{renew: func(renewCtx context.Context, id, owner string, epoch int64, ttl time.Duration) (bool, error) {
			calls++
			if id != "run_1" || owner != "executor_1" || epoch != 3 || ttl != 30*time.Second {
				t.Error("renewal must match acquired credential and TTL")
			}
			deadline, ok := renewCtx.Deadline()
			if !ok || time.Until(deadline) != 10*time.Second {
				t.Error("renewal must have a timeout")
			}
			return true, nil
		}}}
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.renewLease(ctx, "run_1", &execution{cancel: cancel, lease: &lease{Owner: "executor_1", Epoch: 3}}, 30*time.Second)
		}()
		synctest.Wait()
		time.Sleep(10 * time.Second)
		synctest.Wait()
		if calls != 1 || ctx.Err() != nil {
			t.Fatalf("calls=%d, ctx error=%v", calls, ctx.Err())
		}
		cancel()
		<-done
		time.Sleep(20 * time.Second)
		if calls != 1 {
			t.Fatal("renewal continued after cancellation")
		}
	})
}

func TestRenewLeaseFailureCancelsExecution(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "lease lost"},
		{name: "storage error", err: errors.New("storage unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				s := &service{store: &renewalStore{renew: func(context.Context, string, string, int64, time.Duration) (bool, error) {
					return false, tt.err
				}}}
				done := make(chan struct{})
				go func() {
					defer close(done)
					s.renewLease(ctx, "run_1", &execution{cancel: cancel, lease: &lease{}}, 30*time.Second)
				}()
				<-done
				if !errors.Is(ctx.Err(), context.Canceled) {
					t.Fatal("failed renewal did not cancel execution")
				}
			})
		})
	}
}

func TestRenewLeaseTimeoutCancelsExecution(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		s := &service{store: &renewalStore{renew: func(ctx context.Context, _ string, _ string, _ int64, _ time.Duration) (bool, error) {
			<-ctx.Done()
			return false, ctx.Err()
		}}}
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.renewLease(ctx, "run_1", &execution{cancel: cancel, lease: &lease{}}, 30*time.Second)
		}()
		<-done
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatal("timed out renewal did not cancel execution")
		}
	})
}
