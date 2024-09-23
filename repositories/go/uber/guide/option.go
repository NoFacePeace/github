package guide

import "go.uber.org/zap"

type options struct {
	cache  bool
	logger *zap.Logger
}

type Option interface {
	apply(*options)
}

type cacheOption bool

func (c cacheOption) apply(opts *options) {
	opts.cache = bool(c)
}

func WithCache(c bool) Option {
	return cacheOption(c)
}

type loggerOption struct {
	Log *zap.Logger
}

func (l loggerOption) apply(opts *options) {
	opts.logger = l.Log
}

func WithLogger(log *zap.Logger) Option {
	return loggerOption{Log: log}
}
