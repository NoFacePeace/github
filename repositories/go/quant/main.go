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
	cmd  string
	code string
)

func init() {
	flag.StringVar(&cmd, "cmd", "web", "command")
	flag.StringVar(&code, "code", "sh600188", "code")
}

func main() {
	// switch cmd {
	// case "multiple":
	// 	multiple()
	// case "single":
	// 	single(code)
	// }
	multiple()
}

func multiple() {
	stocks, err := finance.ListStocks(finance.WithListStocksCount(500))
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range stocks {
		single(v.Name, v.Code)
	}
}

func single(name, code string) {
	ps, err := indicator.AllPrice(code)
	if err != nil {
		log.Fatal(err)
	}
	mn := 1
	mx := 30
	s, l, win, ps := indicator.SMABestCrossPercent(ps, mn, mx, 3)
	cnt := len(ps) / 2
	if cnt == 0 {
		return
	}
	if len(ps)%2 == 0 {
		return
	}
	date := ps[len(ps)-1].Date.Format(datetime.LayoutDateWithDash)
	if date != "2024-09-30" {
		return
	}
	if win < 70 {
		return
	}
	fmt.Println(name, " ", date, " buy:", code, s, l, cnt, win)
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

func SMACrossDetail(code string, s, l int) {
	ps, err := indicator.AllPrice(code)
	if err != nil {
		log.Fatal(err)
	}
	ps = indicator.Cross(ps, indicator.SMA(ps, s), indicator.SMA(ps, l))
	for i := 0; i < len(ps)-1; i += 2 {
		fmt.Println(
			ps[i].Date.Format(datetime.LayoutDateWithLine),
			ps[i].Price,
			ps[i+1].Date.Format(datetime.LayoutDateWithLine),
			ps[i+1].Price,
			ps[i+1].Price-ps[i].Price,
			(ps[i+1].Price-ps[i].Price)/ps[i].Price*100,
		)
	}
	fmt.Println("count: ", len(ps)/2, "win: ", indicator.Win(ps))
}
