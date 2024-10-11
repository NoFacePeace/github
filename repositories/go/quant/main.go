package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"sync"
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
	var wg sync.WaitGroup
	for _, v := range stocks {
		wg.Add(1)
		go func(s finance.Stock) {
			single(s.Name, s.Code)
			wg.Done()
		}(v)
	}
	wg.Wait()
}

func single(name, code string) {
	points, err := indicator.AllPrice(code)
	if err != nil {
		log.Fatal(err)
	}
	mn := 1
	mx := 30
	// s, l, win, ps := indicator.SMABestCrossMaxPercent(ps, mn, mx, 3)
	s, l, win, ps := indicator.SMAGoldenCrossLastPercent(points, mn, mx, 7, 2)
	cnt := len(ps) / 2
	if win >= 85 && len(ps)%2 > 0 {
		date := ps[len(ps)-1].Date.Format(datetime.LayoutDateWithDash)
		fmt.Println(name, code, date, "golden", s, l, cnt, win)
	}
	s, l, win, ps = indicator.SMADeadCrossLastPercent(points, mn, mx, 7, 2)
	cnt = len(ps) / 2
	if win >= 85 && len(ps) > 0 {
		date := ps[len(ps)-1].Date.Format(datetime.LayoutDateWithDash)
		fmt.Println(name, code, date, "dead", s, l, cnt, win)
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
