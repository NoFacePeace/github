package manager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type Manager interface {
	// 启动管理器
	Start(context.Context) error
	// 添加任务
	Add(Runnable) error
}

func New() (Manager, error) {
	errChan := make(chan error, 1)
	runnables := newRunnables(errChan)
	return &controllerManager{
		errChan:   errChan,
		runnables: runnables,
	}, nil
}

type controllerManager struct {
	// 锁
	sync.Mutex
	// error channel
	errChan   chan error
	runnables *runnables

	started bool

	internalCtx    context.Context
	internalCancel context.CancelFunc

	stopProcedureEngaged *int64

	gracefulShutdownTimeout time.Duration
	shutdownCtx             context.Context
	internalProceduresStop  chan struct{}
}

func (cm *controllerManager) Start(ctx context.Context) (err error) {
	// 加锁，限制并发
	cm.Lock()
	// 防止重复启动
	if cm.started {
		cm.Unlock()
		return errors.New("manager already started")
	}
	cm.started = true

	// start 会一直阻塞，如果启动失败，需要解锁
	var ready bool
	defer func() {
		if !ready {
			cm.Unlock()
		}
	}()

	// 初始化内部 context
	cm.internalCtx, cm.internalCancel = context.WithCancel(ctx)

	// 停止
	stopComplete := make(chan struct{})
	defer close(stopComplete)
	defer func() {
		stopErr := cm.engageStopProcedure(stopComplete)
		if stopErr != nil {
			if err != nil {
				err = errors.Join(err, stopErr)
			} else {
				err = stopErr
			}
		}
	}()

	if err := cm.runnables.Others.Start(cm.internalCtx); err != nil {
		return fmt.Errorf("failed to start to other runnables: [%w]", err)
	}
	// ready 成功，解锁
	ready = true
	cm.Unlock()
	slog.Info("manager started")
	select {
	case <-ctx.Done():
		slog.Info("manager stopped")
		return nil
	case err := <-cm.errChan:
		return err
	}
}

func (cm *controllerManager) Add(r Runnable) error {
	// 加锁，限制并发
	cm.Lock()
	defer cm.Unlock()
	return cm.add(r)
}

func (cm *controllerManager) add(r Runnable) error {
	// 任务添加到任务集
	return cm.runnables.Add(r)
}

func (cm *controllerManager) engageStopProcedure(stopComplete <-chan struct{}) error {
	if !atomic.CompareAndSwapInt64(cm.stopProcedureEngaged, 0, 1) {
		return errors.New("stop procedure already engaged")
	}
	var shutdownCancel context.CancelFunc
	if cm.gracefulShutdownTimeout < 0 {
		cm.shutdownCtx, shutdownCancel = context.WithCancel(context.Background())
	} else {
		cm.shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), cm.gracefulShutdownTimeout)
	}
	defer shutdownCancel()

	var closeOnce sync.Once
	go func() {
		for {
			closeOnce.Do(func() {
				close(cm.internalProceduresStop)
				cm.internalCancel()
			})
			select {
			case err, ok := <-cm.errChan:
				if ok {
					slog.Error("error", "error", err)
				}
			case <-stopComplete:
				return
			}
		}
	}()

	go func() {
		cm.runnables.Others.StopAndWait(cm.shutdownCtx)
		shutdownCancel()
	}()
	<-cm.shutdownCtx.Done()
	if err := cm.shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		if errors.Is(err, context.DeadlineExceeded) {
			if cm.gracefulShutdownTimeout > 0 {
				return fmt.Errorf("failed waiting for all runnables to end within grace period of %s: %w", cm.gracefulShutdownTimeout, err)
			}
			return nil
		}
		return err
	}
	return nil
}

// runnables 任务集，包含各种任务组
type runnables struct {
	Others *runnableGroup
}

func newRunnables(errChan chan error) *runnables {
	return &runnables{
		Others: newRunnableGroup(errChan),
	}
}

func (r *runnables) Add(fn Runnable) error {
	switch runnale := fn.(type) {
	default:
		return r.Others.Add(runnale, nil)
	}
}
