package goroutine

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

var DefaultHandler = func(r any) {
	slog.Error("goroutine recover panic", "error", fmt.Sprint(r), "stack", string(debug.Stack()))
}

type options struct {
	handler func(r any)
	loop    bool
}

type Option func(*options)

func Recover(f func(), opts ...Option) {
	options := &options{
		handler: DefaultHandler,
	}
	for _, opt := range opts {
		opt(options)
	}
	w := func() (success bool) {
		success = false
		defer func() {
			if r := recover(); r != nil {
				options.handler(r)
			}
		}()
		f()
		success = true
		return
	}
	go func() {
		for {
			if w() {
				break
			}
			if !options.loop {
				break
			}
		}
	}()
}

func WithHandler(handler func(r any)) Option {
	return func(opts *options) {
		opts.handler = handler
	}
}

func WithLoop() Option {
	return func(opts *options) {
		opts.loop = true
	}
}
