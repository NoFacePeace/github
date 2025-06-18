package main

import (
	"net/http"

	"github.com/NoFacePeace/github/repositories/go/utils/prometheus"
	"github.com/NoFacePeace/github/repositories/go/utils/signal"
)

func main() {
	ctx := signal.SetupSignalHandler()
	prometheus.Start()
	http.ListenAndServe(":80", nil)
	<-ctx.Done()
}
