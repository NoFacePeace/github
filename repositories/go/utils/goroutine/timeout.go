package goroutine

import (
	"errors"
	"time"
)

var (
	errTimeout = errors.New("function timeout")
)

func GoWithTimeout(fn func() error, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- fn()
	}()
	select {
	case err := <-errCh:
		return err
	case <-time.After(timeout):
		return errTimeout
	}
}
