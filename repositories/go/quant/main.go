package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/NoFacePeace/github/repositories/go/external/tencnet/finance"
	"github.com/NoFacePeace/github/repositories/go/quant/grafana"
	"github.com/NoFacePeace/github/repositories/go/quant/indicator"
)

var (
	cmd string
)

func init() {
	flag.StringVar(&cmd, "cmd", "", "command")
}

func main() {

	// test()
	// single("sh600941")
	// multiple()
	web()
}

func test() {
	code := "sh600941"
	ps, err := indicator.AllPrice(code)
	if err != nil {
		log.Fatal(err)
	}
	ps = indicator.Cross(ps, indicator.SMA(ps, 14), indicator.SMA(ps, 18))
	for i := 0; i < len(ps)-1; i += 2 {
		fmt.Println(ps[i], ps[i+1])
	}
	fmt.Println(indicator.Win(ps), indicator.Earn(ps), indicator.Money(ps))
}

func single(code string) {
	ps, err := indicator.Price(code)
	if err != nil {
		log.Fatal(err)
	}
	mn := 2
	mx := 30
	s, l, win, ps := indicator.SMABestCross(ps, mn, mx)
	fmt.Println(code, s, l, win, len(ps), ps[len(ps)-1])
}

func multiple() {
	stocks, err := finance.ListStocks(finance.AStock, finance.WithCount(200))
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range stocks {
		fmt.Print(v.Name)
		single(v.Code)
	}
}

func web() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	g := grafana.New()
	g.Start()
	log.Println("start")
	<-ctx.Done()
	stop()
	g.Stop()
	log.Println("stop")
}
