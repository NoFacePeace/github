package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/NoFacePeace/github/repositories/go/external/tencent/finance"
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
	stocks, err := finance.ListStocks(finance.WithListStocksCount(1000))
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
	points, err := indicator.Price(code)
	if err != nil {
		log.Fatal(err)
	}
	mn := 1
	mx := 30
	// SMACrossMaxPercentBest(name, code, points, mn, mx, 3)
	// SMAGoldenCross(name, code, points, mn, mx)
	SMACrossLastPercentBest(name, code, points, mn, mx, 4, 2)
}
func SMACrossLastPercentBest(name, code string, points []indicator.Point, mn, mx, last int, percent float64) {
	s, l, win, ps := indicator.SMAGoldenCrossLastPercent(points, mn, mx, last, percent)
	n := len(ps)
	if n <= 0 {
		return
	}
	date := ps[n-1].Date.Format(datetime.LayoutDateWithDash)
	price := ps[n-1].Price
	cnt := n / 2
	if n%2 == 0 {
		fmt.Println(name, code, date, "sell", "golden", price, price*1.02, s, l, cnt, win)
	} else {
		fmt.Println(name, code, date, "buy", "golden", price, price*1.02, s, l, cnt, win)
	}
	s, l, win, ps = indicator.SMADeadCrossLastPercent(points, mn, mx, last, percent)
	n = len(ps)
	if n <= 0 {
		return
	}
	date = ps[n-1].Date.Format(datetime.LayoutDateWithDash)
	price = ps[n-1].Price
	cnt = n / 2
	if n%2 == 0 {
		fmt.Println(name, code, date, "sell", "dead", price, price*1.02, s, l, cnt, win)
	} else {
		fmt.Println(name, code, date, "buy", "dead", price, price*1.02, s, l, cnt, win)
	}
}

func SMACrossMaxPercentBest(name, code string, points []indicator.Point, mn, mx int, percent float64) {
	s, l, win, ps := indicator.SMACrossMaxPercentBest(points, mn, mx, percent)
	n := len(ps)
	if n == 0 {
		fmt.Println(name, code)
		return
	}
	cnt := len(ps) / 2
	date := ps[len(ps)-1].Date.Format(datetime.LayoutDateWithDash)
	price := ps[len(ps)-1].Price
	if n%2 == 0 {
		fmt.Println(name, code, date, "sell", price, s, l, cnt, win)
	} else {
		fmt.Println(name, code, date, "buy", price, s, l, cnt, win)
	}
}

func SMAGoldenCross(name, code string, points []indicator.Point, mn, mx int) {
	s, l, win, ps := indicator.SMAGoldenCrossBest(points, mn, mx)
	n := len(ps)
	if n == 0 {
		fmt.Println(name, code)
		return
	}
	cnt := len(ps) / 2
	date := ps[len(ps)-1].Date.Format(datetime.LayoutDateWithDash)
	if len(ps)%2 == 0 {
		fmt.Println(name, code, date, "sell", s, l, cnt, win)
	} else {
		fmt.Println(name, code, date, "buy", s, l, cnt, win)
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
	ps = indicator.GoldenCross(ps, indicator.SMA(ps, s), indicator.SMA(ps, l))
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
