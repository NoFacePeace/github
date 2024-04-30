package tencent

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/NoFacePeace/github/repositories/go/util/datetime"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	// board type
	BoardType1 = "hy1"
	BoardType2 = "hy2"
	// board code
	BoardCode1         = "aStock"
	LayoutDateWithLine = "2006-01-02"
)

type Tencent struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Tencent {
	err := db.Set("gorm:table_options", "ENGINE=ReplacingMergeTree PRIMARY KEY (name,date,code)").AutoMigrate(&Kline{})
	if err != nil {
		slog.Error(err.Error())
	}
	err = db.Set("gorm:table_options", "ENGINE=ReplacingMergeTree PRIMARY KEY (name,date, code)").AutoMigrate(&Plate{})
	if err != nil {
		slog.Error(err.Error())
	}
	err = db.Set("gorm:table_options", "ENGINE=ReplacingMergeTree PRIMARY KEY (name, date, code)").AutoMigrate(&Stock{})
	if err != nil {
		slog.Error(err.Error())
	}
	err = db.Set("gorm:table_options", "ENGINE=ReplacingMergeTree PRIMARY KEY (plate_name, stock_name, plate_code,  stock_code)").AutoMigrate(&StockPlate{})
	if err != nil {
		slog.Error(err.Error())
	}
	return &Tencent{
		db: db,
	}
}

func (t *Tencent) Daily() {
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

func (t *Tencent) Weekly() {
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

func (t *Tencent) History() {
	slog.Info("start to scape history")
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

func (t *Tencent) ScrapePlateBoard(board string) ([]Plate, error) {
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
func (t *Tencent) ScrapeStockBoard(code string) ([]Stock, error) {
	arr, err := getFullBoardRankList(code)
	if tx := t.db.Create(arr); tx.Error != nil {
		return nil, errors.New(tx.Error.Error())
	}
	return arr, err
}
func (t *Tencent) ScrapeKline(code string) error {
	arr, err := getKline(code, 10, "day", "")
	if err != nil {
		return err
	}
	tx := t.db.Create(arr)
	if tx.Error != nil {
		return errors.New(tx.Error.Error())
	}
	return nil
}

func (t *Tencent) ScrapeKlineHistory(code string) error {
	limit := 670
	toDate := ""
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
		toDate = arr[0].Date.Format(LayoutDateWithLine)
		slog.Info(toDate)
	}
	return nil
}

func getFullRank(boardType string) ([]Plate, error) {
	offset := 0
	count := 40
	arr := []Plate{}
	for {
		tmp, err := getRank(boardType, offset, count)
		if err != nil {
			return nil, err
		}
		arr = append(arr, tmp...)
		if len(tmp) < count {
			break
		}
		offset += count
	}

	return arr, nil
}

func getRank(boardType string, offset int, count int) ([]Plate, error) {
	url := fmt.Sprintf("https://proxy.finance.qq.com/cgi/cgi-bin/rank/pt/getRank?count=%v&offset=%v&direct=down&board_type=%s&sort_type=PriceRatio", count, offset, boardType)
	var resp GetRankResp
	if err := getBody(url, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Msg)
	}
	arr := []Plate{}
	for _, v := range resp.Data.RankList {
		b := Plate{}
		b.Code = v.Code
		b.Name = v.Name
		b.Zxj, _ = strconv.ParseFloat(v.Zxj, 64)
		b.Zdf, _ = strconv.ParseFloat(v.Zdf, 64)
		b.Zd, _ = strconv.ParseFloat(v.Zd, 64)
		b.Hsl, _ = strconv.ParseFloat(v.Hsl, 64)
		b.Lb, _ = strconv.ParseFloat(v.Lb, 64)
		b.Volume, _ = strconv.ParseFloat(v.Volume, 64)
		b.Turnover, _ = strconv.ParseFloat(v.Turnover, 64)
		b.Zsz, _ = strconv.ParseFloat(v.Zsz, 64)
		b.Ltsz, _ = strconv.ParseFloat(v.Ltsz, 64)
		b.Speed, _ = strconv.ParseFloat(v.Speed, 64)
		b.ZdfD5, _ = strconv.ParseFloat(v.ZdfD5, 64)
		b.ZdfD20, _ = strconv.ParseFloat(v.ZdfD20, 64)
		b.ZdfD60, _ = strconv.ParseFloat(v.ZdfD60, 64)
		b.ZdfY, _ = strconv.ParseFloat(v.ZdfY, 64)
		b.ZdfW52, _ = strconv.ParseFloat(v.ZdfW52, 64)
		b.Zllr, _ = strconv.ParseFloat(v.Zllr, 64)
		b.Zllc, _ = strconv.ParseFloat(v.Zllc, 64)
		b.Zljlr, _ = strconv.ParseFloat(v.Zljlr, 64)
		b.ZljlrD5, _ = strconv.ParseFloat(v.ZljlrD5, 64)
		b.ZljlrD20, _ = strconv.ParseFloat(v.ZljlrD20, 64)
		b.Zgb = v.Zgb
		b.Lzg.Code = v.Lzg.Code
		b.Lzg.Name = v.Lzg.Name
		b.Lzg.Zxj, _ = strconv.ParseFloat(v.Lzg.Zxj, 64)
		b.Lzg.Zdf, _ = strconv.ParseFloat(v.Lzg.Zdf, 64)
		b.Lzg.Zd, _ = strconv.ParseFloat(v.Lzg.Zd, 64)
		b.StockType = v.StockType
		b.Date = time.Now().Local()
		arr = append(arr, b)
	}
	return arr, nil
}

type GetRankResp struct {
	Code int `json:"code"`
	Data struct {
		RankList []struct {
			Code     string `json:"code"`
			Name     string `json:"name"`
			Zxj      string `json:"zxj"`
			Zdf      string `json:"zdf"`
			Zd       string `json:"zd"`
			Hsl      string `json:"hsl"`
			Lb       string `json:"lb"`
			Volume   string `json:"volume"`
			Turnover string `json:"turnover"`
			Zsz      string `json:"zsz"`
			Ltsz     string `json:"ltsz"`
			Speed    string `json:"speed"`
			ZdfD5    string `json:"zdf_d5"`
			ZdfD20   string `json:"zdf_d20"`
			ZdfD60   string `json:"zdf_d60"`
			ZdfY     string `json:"zdf_y"`
			ZdfW52   string `json:"zdf_w52"`
			Zllr     string `json:"zllr"`
			Zllc     string `json:"zllc"`
			Zljlr    string `json:"zljlr"`
			ZljlrD5  string `json:"zljlr_d5"`
			ZljlrD20 string `json:"zljlr_d20"`
			Zgb      string `json:"zgb"`
			Lzg      struct {
				Code string `json:"code"`
				Name string `json:"name"`
				Zxj  string `json:"zxj"`
				Zdf  string `json:"zdf"`
				Zd   string `json:"zd"`
			} `json:"lzg"`
			StockType string `json:"stock_type"`
		} `json:"rank_list"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"data"`
	Msg string `json:"msg"`
}
type Plate struct {
	Code     string
	Name     string
	Zxj      float64 `gorm:"comment:最新价"`
	Zdf      float64 `gorm:"comment:涨跌幅"`
	Zd       float64 `gorm:"comment:涨跌"`
	Hsl      float64 `gorm:"comment:换手率"`
	Lb       float64 `gorm:"comment:量比"`
	Volume   float64 `gorm:"comment:成交量，单位万"`
	Turnover float64 `gorm:"comment:成交额，单位万"`
	Zsz      float64 `gorm:"comment:总市值，单位万"`
	Ltsz     float64 `gorm:"comment:流通市值，单位万"`
	Speed    float64 `gorm:"comment:5 分钟涨速"`
	ZdfD5    float64 `gorm:"comment:5 日涨跌幅"`
	ZdfD20   float64 `gorm:"comment:20 日涨跌幅"`
	ZdfD60   float64 `gorm:"comment:60 日 涨跌幅"`
	ZdfY     float64 `gorm:"comment:年涨跌幅"`
	ZdfW52   float64 `gorm:"comment:52 周涨跌幅"`
	Zllr     float64 `gorm:"comment:主力流入"`
	Zllc     float64 `gorm:"comment:主力流出"`
	Zljlr    float64 `gorm:"comment:主力净流入"`
	ZljlrD5  float64 `gorm:"comment:5 日主力净流入"`
	ZljlrD20 float64 `gorm:"comment:20 日主力净流入"`
	Zgb      string
	Lzg      struct {
		Code string
		Name string
		Zxj  float64 `gorm:"comment:最新价"`
		Zdf  float64 `gorm:"comment:涨跌幅"`
		Zd   float64 `gorm:"comment:涨幅"`
	} `gorm:"embedded;embeddedPrefix:lzg_"`
	StockType string    `gorm:"comment:类型"`
	Date      time.Time `gorm:"type:date"`
}

func getFullBoardRankList(code string) ([]Stock, error) {
	offset := 0
	count := 40
	arr := []Stock{}
	for {
		tmp, err := getBoardRankList(code, offset, count)
		if err != nil {
			return nil, err
		}
		arr = append(arr, tmp...)
		if len(tmp) < count {
			break
		}
		offset += count
	}
	return arr, nil
}

func getBoardRankList(code string, offset, count int) ([]Stock, error) {
	url := fmt.Sprintf(
		"https://proxy.finance.qq.com/cgi/cgi-bin/rank/hs/getBoardRankList?board_code=%s&sort_type=priceRatio&direct=down&offset=%v&count=%v&_appver=11.14.0",
		code,
		offset,
		count,
	)
	var resp GetBoardRankListResp
	if err := getBody(url, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Msg)
	}
	arr := []Stock{}
	for _, v := range resp.Data.RankList {
		s := Stock{}
		s.InStock = v.InStock
		s.Date = time.Now().Local()
		arr = append(arr, s)
	}
	return arr, nil
}

type Stock struct {
	InStock
	Date time.Time `gorm:"type:date"`
}
type GetBoardRankListResp struct {
	Code int `json:"code"`
	Data struct {
		RankList []struct {
			InStock
			Labels []struct {
				Label int    `json:"label"`
				Value []any  `json:"value"`
				Name  string `json:"name"`
			} `json:"labels,omitempty"`
		} `json:"rank_list"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type InStock struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Zxj       float64 `json:"zxj,string" gorm:"comment:最新价"`
	Zdf       float64 `json:"zdf,string" gorm:"comment:涨跌幅"`
	Zd        float64 `json:"zd,string" gorm:"comment:涨跌额"`
	Hsl       float64 `json:"hsl,string" gorm:"comment:换手率"`
	Lb        float64 `json:"lb,string" gorm:"comment:量比"`
	Zf        float64 `json:"zf,string" gorm:"comment:振幅"`
	Volume    float64 `json:"volume,string" gorm:"comment:成交量"`
	Turnover  float64 `json:"turnover,string" gorm:"comment:成交额"`
	PeTtm     float64 `json:"pe_ttm,string" gorm:"comment:市盈 TTM"`
	Pn        float64 `json:"pn,string" gorm:"comment:市净率"`
	Zsz       float64 `json:"zsz,string" gorm:"comment:总市值"`
	Ltsz      float64 `json:"ltsz,string" gorm:"comment:流通市值"`
	State     string  `json:"state"`
	Speed     float64 `json:"speed,string" gorm:"comment:5分钟涨速"`
	ZdfY      float64 `json:"zdf_y,string" gorm:"comment:年初至今涨跌幅"`
	ZdfD5     float64 `json:"zdf_d5,string" gorm:"comment:5日涨跌幅"`
	ZdfD10    float64 `json:"zdf_d10,string" gorm:"comment:10涨跌幅"`
	ZdfD20    float64 `json:"zdf_d20,string" gorm:"comment:20日涨跌幅"`
	ZdfD60    float64 `json:"zdf_d60,string" gorm:"comment:60日涨跌幅"`
	ZdfW52    float64 `json:"zdf_w52,string" gorm:"comment:52周涨跌幅"`
	Zljlr     float64 `json:"zljlr,string" gorm:"comment:主力净流入"`
	Zllr      float64 `json:"zllr,string" gorm:"comment:主力流入"`
	Zllc      float64 `json:"zllc,string" gorm:"comment:主力流出"`
	ZllrD5    float64 `json:"zllr_d5,string" gorm:"coment:5日主力流入"`
	ZllcD5    float64 `json:"zllc_d5,string" gorm:"coment:5日主力流出"`
	StockType string  `json:"stock_type" gorm:"股票类型"`
}

func getKline(code string, limit int, ktype string, toDate string) ([]Kline, error) {
	url := fmt.Sprintf("https://proxy.finance.qq.com/cgi/cgi-bin/stockinfoquery/kline/app/get?ktype=%v&limit=%v&code=%v&toDate=%v", ktype, limit, code, toDate)
	var resp GetKlineResp
	err := getBody(url, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}
	arr := []Kline{}
	for _, v := range resp.Data.Nodes {
		line := Kline{}
		line.Open, _ = strconv.ParseFloat(v.Open, 64)
		line.Last, _ = strconv.ParseFloat(v.Last, 64)
		line.High, _ = strconv.ParseFloat(v.High, 64)
		line.Low, _ = strconv.ParseFloat(v.Low, 64)
		line.Volume, _ = strconv.ParseFloat(v.Volume, 64)
		line.Amount, _ = strconv.ParseFloat(v.Amount, 64)
		line.Exchange, _ = strconv.ParseFloat(v.Exchange, 64)
		line.ExchangeRaw, _ = strconv.ParseFloat(v.ExchangeRaw, 64)
		line.Oi, _ = strconv.ParseFloat(v.Oi, 64)
		line.TradeDays, _ = strconv.ParseFloat(v.TradeDays, 64)
		line.Dividend = v.Dividend
		line.AddZdf = v.AddZdf
		line.Date, _ = time.Parse(LayoutDateWithLine, v.Date)
		line.Code = code
		line.Name = resp.Data.Qt.Fields[1]
		arr = append(arr, line)
	}
	return arr, nil
}

type GetKlineResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		StockCode string `json:"stockCode"`
		Nodes     []struct {
			Open        string `json:"open"`
			Last        string `json:"last"`
			High        string `json:"high"`
			Low         string `json:"low"`
			Volume      string `json:"volume"`
			Amount      string `json:"amount"`
			Exchange    string `json:"exchange"`
			ExchangeRaw string `json:"exchangeRaw"`
			Date        string `json:"date"`
			Oi          string `json:"oi"`
			TradeDays   string `json:"tradeDays"`
			Dividend    string `json:"dividend"`
			AddZdf      string `json:"addZdf"`
			Finance     any    `json:"finance"`
		} `json:"nodes"`
		Qt struct {
			Fields []string `json:"fields"`
			Market string   `json:"market"`
			Zhishu []string `json:"zhishu"`
		} `json:"qt"`
		FundQt      any    `json:"fundQt"`
		CvbondQt    any    `json:"cvbondQt"`
		Prec        string `json:"prec"`
		FsStartDate string `json:"fsStartDate"`
		Pandata     any    `json:"pandata"`
		Introduce   string `json:"introduce"`
		Funddata    any    `json:"funddata"`
		WarrantInfo any    `json:"warrantInfo"`
		Fqtype      string `json:"fqtype"`
		Attribute   any    `json:"attribute"`
		Fs          any    `json:"fs"`
		OpPoints    []any  `json:"opPoints"`
	} `json:"data"`
}

type Kline struct {
	Code        string
	Name        string
	Open        float64 `gorm:"comment:开盘价"`
	Last        float64 `gorm:"comment:最新价"`
	High        float64 `gorm:"comment:最高价"`
	Low         float64 `gorm:"comment:最低价"`
	Volume      float64 `gorm:"comment:成交量"`
	Amount      float64 `gorm:"comment:成交额"`
	Exchange    float64 `gorm:"comment:换手率"`
	ExchangeRaw float64
	Date        time.Time `gorm:"type:date;comment:日期"`
	Oi          float64
	TradeDays   float64
	Dividend    string
	AddZdf      string
}

func getPlate(code string) ([]StockPlate, error) {
	url := fmt.Sprintf(
		"https://proxy.finance.qq.com/ifzqgtimg/appstock/app/stockinfo/plate?code=%s&_appver=11.14.0",
		code,
	)
	var resp GetPlateResp
	if err := getBody(url, &resp); err != nil {
		return nil, nil
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Msg)
	}
	arr := []StockPlate{}
	for _, v := range resp.Data.Plate {
		tmp := StockPlate{}
		tmp.StockCode = code
		tmp.PlateCode = v.ID
		tmp.PlateName = v.Name
		arr = append(arr, tmp)
	}
	for _, v := range resp.Data.Concept {
		tmp := StockPlate{}
		tmp.StockCode = code
		tmp.PlateCode = v.ID
		tmp.PlateName = v.Name
		arr = append(arr, tmp)
	}
	for _, v := range resp.Data.Area {
		tmp := StockPlate{}
		tmp.StockCode = code
		tmp.PlateCode = v.ID
		tmp.PlateName = v.Name
		arr = append(arr, tmp)
	}
	return arr, nil
}

type GetPlateResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Plate []struct {
			Name  string `json:"name"`
			ID    string `json:"id"`
			Level string `json:"level"`
		} `json:"plate"`
		Concept []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
			Tag  string `json:"tag"`
		} `json:"concept"`
		Area []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"area"`
	} `json:"data"`
}

type StockPlate struct {
	StockCode string
	StockName string
	PlateCode string
	PlateName string
	Date      time.Time `gorm:"type:date"`
}

func getBody(url string, rsp any) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.New(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(err.Error())
	}
	err = json.Unmarshal(body, rsp)
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}
