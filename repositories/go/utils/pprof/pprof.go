package pprof

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

const (
	defaultPort = 6060
)

func Start(options ...Option) {
	cfg := newConfig()
	for _, option := range options {
		option.apply(cfg)
	}
	go func() {
		log.Println(http.ListenAndServe(":"+strconv.Itoa(cfg.port), nil))
	}()
}

type config struct {
	port int
}

func newConfig() *config {
	return &config{
		port: defaultPort,
	}
}

type Option interface {
	apply(*config)
}

func WithPort(port int) Option {
	return portOption(port)
}

type portOption int

func (p portOption) apply(cfg *config) {
	cfg.port = int(p)
}
