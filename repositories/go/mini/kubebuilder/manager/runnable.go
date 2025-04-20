package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Runnable interface {
	Start(context.Context) error
}

type readyRunnable struct {
	Runnable
	// 检测 runnable 是否 start 成功，阻塞
	Check runnableCheck
	// ready 标志
	signalReady bool
}

type runnableCheck func(ctx context.Context) bool

type runnableGroup struct {
	ctx    context.Context
	cancel context.CancelFunc

	// 启动锁
	start sync.Mutex
	// 启动标志
	started bool
	// 启动队列
	startQueue []*readyRunnable
	// 启动函数执行
	startOnce sync.Once
	// 已经 ready runnable channel
	startReadyCh chan *readyRunnable

	stop     sync.RWMutex
	stopped  bool
	stopOnce sync.Once

	// 还没有 ready runnable channal
	ch chan *readyRunnable

	errChan chan error
	wg      *sync.WaitGroup
}

func newRunnableGroup(errChan chan error) *runnableGroup {
	r := &runnableGroup{
		startReadyCh: make(chan *readyRunnable),
		errChan:      errChan,
		ch:           make(chan *readyRunnable),
		wg:           new(sync.WaitGroup),
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	return r
}

func (r *runnableGroup) Started() bool {
	r.start.Lock()
	defer r.start.Unlock()
	return r.started
}

func (r *runnableGroup) Add(rn Runnable, ready runnableCheck) error {
	// check if stop
	r.stop.RLock()
	if r.stopped {
		r.stop.RUnlock()
		return fmt.Errorf("runnable group stop")
	}
	r.stop.RUnlock()

	if ready == nil {
		ready = func(ctx context.Context) bool { return true }
	}
	readyRunnable := &readyRunnable{
		Runnable: rn,
		Check:    ready,
	}

	// if don't start, add queue
	r.start.Lock()
	if !r.started {
		r.startQueue = append(r.startQueue, readyRunnable)
		r.start.Unlock()
		return nil
	}
	r.start.Unlock()

	r.stop.RLock()
	defer r.stop.RUnlock()
	if r.stopped {
		return fmt.Errorf("runnable group stop")
	}
	r.ch <- readyRunnable
	return nil
}

// 启动 reconcile
// 遍历 start queue，设置为 ready，发送到 channel
// reconcile 会从 channel 获取并执行 start
func (r *runnableGroup) Start(ctx context.Context) (retError error) {
	var retErr error
	r.startOnce.Do(func() {
		defer close(r.startReadyCh)

		go r.reconcile()

		r.start.Lock()
		r.started = true
		for _, rn := range r.startQueue {
			rn.signalReady = true
			r.ch <- rn
		}
		r.start.Unlock()

		if len(r.startQueue) == 0 {
			return
		}

		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); !errors.Is(err, context.Canceled) {
					retErr = err
				}
			case rn := <-r.startReadyCh:
				for i, existing := range r.startQueue {
					if existing == rn {
						r.startQueue = append(r.startQueue[:i], r.startQueue[i+1:]...)
						break
					}
				}
				if len(r.startQueue) == 0 {
					return
				}
			}
		}
	})
	return retErr
}

// 从 channel 获取 runnable

func (r *runnableGroup) reconcile() {
	for runable := range r.ch {
		{
			r.stop.RLock()
			if r.stopped {
				r.errChan <- fmt.Errorf("runnable group stopped")
				r.stop.RUnlock()
				continue
			}
			r.wg.Add(1)
			r.stop.RUnlock()
		}
		go func(rn *readyRunnable) {
			go func() {
				if rn.Check(r.ctx) {
					if rn.signalReady {
						r.startReadyCh <- rn
					}
				}
			}()
			defer r.wg.Done()
			if err := rn.Start(r.ctx); err != nil {
				r.errChan <- err
			}
		}(runable)
	}
}

func (r *runnableGroup) StopAndWait(ctx context.Context) {
	r.stopOnce.Do(func() {
		defer func() {
			r.stop.Lock()
			close(r.ch)
			r.stop.Unlock()
		}()

		_ = r.Start(ctx)
		r.stop.Lock()
		r.stopped = true
		r.cancel()
		done := make(chan struct{})
		go func() {
			defer close(done)
			r.wg.Wait()
		}()

		select {
		case <-done:
		case <-ctx.Done():
		}
	})
}
