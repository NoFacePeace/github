package goroutine

import (
	"sync"
	"testing"
)

func TestRecover(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	Recover(func() {
		panic("test panic")
	}, WithHandler(func(r any) {
		DefaultHandler(r)
		wg.Done()
	}))
	wg.Wait()
	t.Log("No panic occurred")
}

func TestWithLoop(t *testing.T) {
	opts := &options{}
	if opts.loop {
		t.Error("options loop should be true")
	}
	opt := WithLoop()
	opt.apply(opts)
	if !opts.loop {
		t.Error("options loop should be true")
	}
}
