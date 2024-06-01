package tencent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Clickhouse struct {
	db *gorm.DB
}

func NewClickhouse(db *gorm.DB) *Clickhouse {
	return &Clickhouse{
		db: db,
	}
}

func (t *Clickhouse) InitClickHouse() {
	err := t.db.Set("gorm:table_options", "ENGINE=MergeTree PRIMARY KEY (name,date,code)").AutoMigrate(&Kline{})
	if err != nil {
		slog.Error(err.Error())
	}
	err = t.db.Set("gorm:table_options", "ENGINE=MergeTree PRIMARY KEY (name,date, code)").AutoMigrate(&Plate{})
	if err != nil {
		slog.Error(err.Error())
	}
	err = t.db.Set("gorm:table_options", "ENGINE=MergeTree PRIMARY KEY (name, date, code)").AutoMigrate(&Stock{})
	if err != nil {
		slog.Error(err.Error())
	}
	err = t.db.Set("gorm:table_options", "ENGINE=MergeTree PRIMARY KEY (plate_name, stock_name, plate_code,  stock_code)").AutoMigrate(&StockPlate{})
	if err != nil {
		slog.Error(err.Error())
	}
}

func (t *Clickhouse) ScrapePlateBoard(board string) ([]Plate, error) {
	arr, err := getFullRank(board)
	if err != nil {
		return nil, err
	}
	tx := t.db.Create(arr)
	if tx.Error != nil {
		return nil, errors.New(tx.Error.Error())
	}
	return arr, nil
}
func (t *Clickhouse) ScrapeStockBoard(code string) ([]Stock, error) {
	arr, err := getFullBoardRankList(code)
	if tx := t.db.Create(arr); tx.Error != nil {
		return nil, errors.New(tx.Error.Error())
	}
	return arr, err
}
func (t *Clickhouse) ScrapeKline(code string) error {
	arr, err := getKline(code, 1, "day", "")
	if err != nil {
		return err
	}
	tx := t.db.Create(arr)
	if tx.Error != nil {
		return errors.New(tx.Error.Error())
	}
	return nil
}

func (t *Clickhouse) ScrapeKlineHistory(code string) error {
	limit := 670
	toDate := datetime.Yesterday().Format(datetime.LayoutDateWithLine)
	for {
		arr, err := getKline(code, limit, "day", toDate)
		if err != nil {
			return err
		}
		if tx := t.db.Create(arr); tx.Error != nil {
			return errors.New(tx.Error.Error())
		}
		if len(arr) < limit {
			break
		}
		toDate = arr[0].Date.Format(datetime.LayoutDateWithLine)
		slog.Info(code + ": " + toDate)
	}
	return nil
}

func (t *Clickhouse) Daily() {
	slog.Info("start to scrape daily")
	today := time.Now().Local()
	slog.Info(today.GoString())
	// check is weekend
	if datetime.IsWeekend(today) {
		slog.Info("today is weekend")
		return
	}
	// check is holiday
	ok, err := datetime.IsHoliday(today)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	if ok {
		slog.Info("today is holiday")
		return
	}
}

func (t *Clickhouse) Weekly() {
	stocks, err := getFullBoardRankList(BoardCode1)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	for _, v := range stocks {
		time.Sleep(1 * time.Second)
		st, err := getPlate(v.Code)
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			return
		}
		for _, v2 := range st {
			v2.StockName = v.Name
			v2.Date = time.Now().Local()
			if ret := t.db.Create(&v2); ret.Error != nil {
				slog.Error(fmt.Sprintf("%+v", ret.Error))
				return
			}
		}
	}
}
func (t *Clickhouse) History() {
	slog.Info("start to scape history")

}

func (t *Clickhouse) HistoryPlate() {
	plates, err := getFullRank(BoardType2)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	for _, v := range plates {
		time.Sleep(1 * time.Second)
		slog.Info(fmt.Sprintf("start to scrape plate history kline: %v", v.Name))
		err := t.ScrapeKlineHistory(v.Code)
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			continue
		}
		slog.Info(fmt.Sprintf("end to scrape plate history kline: %v", v.Name))
	}
}

func (t *Clickhouse) HistoryStock() {
	stocks, err := getFullBoardRankList(BoardCode1)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	for _, v := range stocks {
		time.Sleep(1 * time.Second)
		slog.Info(fmt.Sprintf("start to scrape stock history kline: %v", v.Name))
		if err := t.ScrapeKlineHistory(v.Code); err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			continue
		}
		slog.Info(fmt.Sprintf("end to scrape stock history kline: %v", v.Name))
	}
	slog.Info("end to scape history")
}

func (t *Clickhouse) DailyPlate() {
	slog.Info("start to scrape plate board")
	arr, err := t.ScrapePlateBoard(BoardType2)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	slog.Info("end to scrape to plate board")
	for _, v := range arr {
		time.Sleep(1 * time.Second)
		slog.Info(fmt.Sprintf("start to scrape to plate kline: %v", v.Name))
		err := t.ScrapeKline(v.Code)
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			continue
		}
		slog.Info(fmt.Sprintf("end to scrape to plate kline: %v", v.Name))
	}
}

func (t *Clickhouse) DailyStock() {
	slog.Info("starting scrape stock board...")
	stock, err := t.ScrapeStockBoard(BoardCode1)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	slog.Info("ending scrape stock board...")
	for _, v := range stock {
		time.Sleep(1 * time.Second)
		slog.Info(fmt.Sprintf("starting scrape stock kline: %v", v.Name))
		if err := t.ScrapeKline(v.Code); err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			return
		}
		slog.Info(fmt.Sprintf("ending scrape stock kline: %v", v.Name))
	}
	slog.Info("end to scrape daily")
}
