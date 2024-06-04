package tencent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
)

type Tencent struct {
	VictoriaMetrics *VictoriaMetrics
	Store           Store
}

func New(store Store) *Tencent {
	return &Tencent{
		VictoriaMetrics: NewVictoriaMetrics("http://localhost:8428/api/v1/import"),
		Store:           store,
	}
}

func (t *Tencent) Daily() {
	if ok := t.IsOpen(); !ok {
		slog.Info("today is not open")
		return
	}
	plates := []string{PlateBoardHyTwo}
	for _, p := range plates {
		slog.Info("start to scrape plate today: " + p)
		t.ScrapePlateToday(p)
		slog.Info("end to scrape plate today: " + p)
	}
	stocks := []string{StockBoardAstock}
	for _, s := range stocks {
		slog.Info("start to scrape stock today: " + s)
		t.ScrapeStockToday(s)
		slog.Info("end to scrape stock today: " + s)
	}
}
func (t *Tencent) History() {
	plates := []string{PlateBoardHyTwo}
	for _, p := range plates {
		slog.Info("start to scrape plate history: " + p)
		t.ScrapePlateHistory(p)
		slog.Info("end to scrape plate history: " + p)
	}
	stocks := []string{StockBoardAstock}
	for _, s := range stocks {
		slog.Info("start to scrape stock history: " + s)
		t.ScrapeStockHistory(s)
		slog.Info("end to scrape stock history: " + s)
	}
}

func (t *Tencent) ScrapePlateToday(typ string) {
	slog.Info("start to scrape plate board today: " + typ)
	plates, err := getFullRank(typ)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	t.Store.SavePlateBoard(plates, map[string]string{
		"type": typ,
	})
	slog.Info("end to scrape plate board today: " + typ)
	for _, p := range plates {
		slog.Info("start to scrape plate kline today: " + p.Name)
		line, err := getKline(p.Code, 1, KlineTypeDay, "")
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
		} else {
			if err := t.Store.SaveKline(line, map[string]string{
				"type":  typ,
				"class": KlineTypeDay,
			}); err != nil {
				slog.Error(fmt.Sprintf("%+v", err))
				return
			}
		}
		slog.Info("end to scrape plate kline today: " + p.Name)
	}
}

func (t *Tencent) ScrapeStockToday(typ string) {
	slog.Info("start to scrape stock board today: " + typ)
	stocks, err := getFullBoardRankList(typ)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	t.Store.SaveStockBoard(stocks, map[string]string{
		"type": typ,
	})
	slog.Info("end to scrape stock board today: " + typ)
	for _, stock := range stocks {
		slog.Info("start to scrape stock kline today: " + stock.Name)
		line, err := getKline(stock.Code, 1, KlineTypeDay, "")
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
		} else {
			if err := t.Store.SaveKline(line, map[string]string{
				"type":  typ,
				"class": KlineTypeDay,
			}); err != nil {
				slog.Error(fmt.Sprintf("%+v", err))
				return
			}
		}
		slog.Info("end to scrape stock kline today: " + stock.Name)
	}
}

func (t *Tencent) ScrapePlateHistory(typ string) {
	plates, err := getFullRank(typ)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	for _, p := range plates {
		slog.Info("start to scrape plate kline history: " + p.Name)
		line, err := getFullKline(p.Code)
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			return
		} else {
			if err := t.Store.SaveKline(line, map[string]string{
				"type":  typ,
				"class": KlineTypeDay,
			}); err != nil {
				slog.Error(fmt.Sprintf("%+v", err))
				return
			}
		}
		slog.Info("end to scrape plate kline history: " + p.Name)
	}
}

func (t *Tencent) ScrapeStockHistory(typ string) {
	stocks, err := getFullBoardRankList(typ)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	for _, stock := range stocks {
		slog.Info("start to scrape stock kline history: " + stock.Name)
		line, err := getFullKline(stock.Code)
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			return
		} else {
			if err := t.Store.SaveKline(line, map[string]string{
				"type":  typ,
				"class": KlineTypeDay,
			}); err != nil {
				slog.Error(fmt.Sprintf("%+v", err))
				return
			}
		}
		slog.Info("end to scrape stock kline history: " + stock.Name)
	}
}

func (t *Tencent) IsOpen() bool {
	today := time.Now().Local()
	slog.Info(today.GoString())
	// check is weekend
	if datetime.IsWeekend(today) {
		slog.Info("today is weekend")
		return false
	}
	// check is holiday
	ok, err := datetime.IsHoliday(today)
	if err != nil {
		slog.Error(err.Error())
		return false
	}
	if ok {
		slog.Info("today is holiday")
		return false
	}
	return true
}
