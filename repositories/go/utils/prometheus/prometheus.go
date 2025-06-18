package prometheus

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Start(options ...Option) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			slog.Error("http listen and server", "error", err)
		}
	}()
}

type Option interface {
}
