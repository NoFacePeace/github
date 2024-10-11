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

type Option interface {
	apply(*options)
}

func Recover(f func(), opts ...Option) {
	options := &options{
		handler: DefaultHandler,
	}
	for _, opt := range opts {
		opt.apply(options)
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
	return &handlerOption{
		handler: handler,
	}
}

func WithLoop() Option {
	return loopOption(true)
}

type handlerOption struct {
	handler func(r any)
}

func (o *handlerOption) apply(opts *options) {
	opts.handler = o.handler
}

type loopOption bool

func (l loopOption) apply(opts *options) {
	opts.loop = bool(l)
}
