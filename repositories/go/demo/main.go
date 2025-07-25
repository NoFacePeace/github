package main

import (
	"github.com/NoFacePeace/github/repositories/go/utils/pprof"
	"github.com/NoFacePeace/github/repositories/go/utils/signal"
)

func main() {
	ctx := signal.SetupSignalHandler()
	pprof.Start(pprof.WithPort(6061))
	<-ctx.Done()
}
