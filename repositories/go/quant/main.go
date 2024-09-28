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
	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
)

var (
	cmd string
)

func init() {
	flag.StringVar(&cmd, "cmd", "web", "command")
}

func main() {

	// test1()
	// single("sh600941")
	multiple()
	// web()
}

func test() {
	code := "sh600941"
	ps, err := indicator.AllPrice(code)
	if err != nil {
		log.Fatal(err)
	}
	ps = indicator.CrossMax(ps, indicator.SMA(ps, 18), indicator.SMA(ps, 27))
	for i := 0; i < len(ps)-1; i += 2 {
		fmt.Println(ps[i], ps[i+1], (ps[i+1].Price-ps[i].Price)/ps[i].Price*100)
	}
	fmt.Println(indicator.Win(ps), indicator.Earn(ps), indicator.Money(ps))
}

func single(code string) {
	ps, err := indicator.AllPrice(code)
	if err != nil {
		log.Fatal(err)
	}
	mn := 1
	mx := 30
	s, l, win, ps := indicator.SMABestCrossMax(ps, mn, mx)
	cnt := len(ps) / 2
	if len(ps)%2 == 0 {
		fmt.Println("sell: ", code, s, l, win, cnt, ps[len(ps)-1])
	} else {
		fmt.Println("buy:", code, s, l, win, cnt, ps[len(ps)-1])
	}
}

func multiple() {
	stocks, err := finance.ListStocks(finance.AStock, finance.WithCount(100))
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

func test1() {
	code := "sh600941"
	ps, err := indicator.AllPrice(code)
	if err != nil {
		log.Fatal(err)
	}
	ps = indicator.CrossMax(ps, indicator.SMA(ps, 18), indicator.SMA(ps, 27))
	for i := 0; i < len(ps)-1; i += 2 {
		fmt.Println(ps[i], ps[i+1], (ps[i+1].Price-ps[i].Price)/ps[i].Price*100)
		fmt.Println(ps[i].Date.Format(datetime.LayoutDateWithLine))
	}
	fmt.Println(indicator.Win(ps), indicator.Earn(ps), indicator.Money(ps))
}
