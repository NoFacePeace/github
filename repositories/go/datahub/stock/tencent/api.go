package tencent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
	"github.com/pkg/errors"
)

var (
	// plate board type
	PlateBoardHyOne = "hy1"
	PlateBoardHyTwo = "hy2"
	// stock board type
	StockBoardAstock = "aStock"

	BoardTypeHyOne = "hy1"
	BoardType2     = "hy2"
	// board code
	BoardCode1   = "aStock"
	KlineTypeDay = "day"
)

func getFullKline(code string) ([]Kline, error) {
	limit := 670
	toDate := datetime.Yesterday().Format(datetime.LayoutDateWithLine)
	arr := []Kline{}
	for {
		tmp, err := getKline(code, limit, "day", toDate)
		if err != nil {
			return nil, err
		}

		arr = append(tmp, arr...)
		if len(tmp) < limit {
			break
		}
		toDate = tmp[0].Date.Format(datetime.LayoutDateWithLine)
	}
	return arr, nil
}

func getFullBoardRankList(code string) ([]Stock, error) {
	offset := 0
	count := 100
	arr := []Stock{}
	for {
		tmp, _, err := getBoardRankList(code, offset, count)
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

func getFullRank(boardType string) ([]Plate, error) {
	offset := 0
	count := 40
	arr := []Plate{}
	for {
		tmp, _, err := getRank(boardType, offset, count)
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

func getBoardRankList(code string, offset, count int) ([]Stock, int, error) {
	url := fmt.Sprintf(
		"https://proxy.finance.qq.com/cgi/cgi-bin/rank/hs/getBoardRankList?board_code=%s&sort_type=priceRatio&direct=down&offset=%v&count=%v&_appver=11.14.0",
		code,
		offset,
		count,
	)
	var resp GetBoardRankListResp
	if err := getBody(url, &resp); err != nil {
		return nil, 0, err
	}
	if resp.Code != 0 {
		return nil, 0, errors.New(resp.Msg)
	}
	arr := []Stock{}
	for _, v := range resp.Data.RankList {
		s := Stock{}
		s.InStock = v.InStock
		s.Date = time.Now()
		arr = append(arr, s)
	}
	return arr, resp.Data.Total, nil
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

func getRank(boardType string, offset int, count int) ([]Plate, int, error) {
	url := fmt.Sprintf("https://proxy.finance.qq.com/cgi/cgi-bin/rank/pt/getRank?count=%v&offset=%v&direct=down&board_type=%s&sort_type=PriceRatio", count, offset, boardType)
	var resp GetRankResp
	if err := getBody(url, &resp); err != nil {
		return nil, 0, err
	}
	if resp.Code != 0 {
		return nil, 0, errors.New(resp.Msg)
	}
	arr := []Plate{}
	for _, v := range resp.Data.RankList {
		b := Plate{}
		b.InnerPlate = v.InnerPlate
		b.Date = time.Now()
		arr = append(arr, b)
	}
	return arr, resp.Data.Total, nil
}

type Plate struct {
	InnerPlate
	Date time.Time `gorm:"type:date"`
}

type InnerPlate struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Zxj      float64 `gorm:"comment:最新价" json:"zxj,string"`
	Zdf      float64 `gorm:"comment:涨跌幅" json:"zdf,string"`
	Zd       float64 `gorm:"comment:涨跌" json:"zd,string"`
	Hsl      float64 `gorm:"comment:换手率" json:"hsl,string"`
	Lb       float64 `gorm:"comment:量比" json:"lb,string"`
	Volume   float64 `gorm:"comment:成交量，单位万" json:"volume,string"`
	Turnover float64 `gorm:"comment:成交额，单位万" json:"turnover,string"`
	Zsz      float64 `gorm:"comment:总市值，单位万" json:"zsz,string"`
	Ltsz     float64 `gorm:"comment:流通市值，单位万" json:"ltsz,string"`
	Speed    float64 `gorm:"comment:5 分钟涨速" json:"speed,string"`
	ZdfD5    float64 `gorm:"comment:5 日涨跌幅" json:"zdf_d5,string"`
	ZdfD20   float64 `gorm:"comment:20 日涨跌幅" json:"zdf_d20,string"`
	ZdfD60   float64 `gorm:"comment:60 日 涨跌幅" json:"zdf_d60,string"`
	ZdfY     float64 `gorm:"comment:年涨跌幅" json:"zdf_y,string"`
	ZdfW52   float64 `gorm:"comment:52 周涨跌幅" json:"zdf_w52,string"`
	Zllr     float64 `gorm:"comment:主力流入" json:"zllr,string"`
	Zllc     float64 `gorm:"comment:主力流出" json:"zllc,string"`
	Zljlr    float64 `gorm:"comment:主力净流入" json:"zljlr,string"`
	ZljlrD5  float64 `gorm:"comment:5 日主力净流入" json:"zljlr_d5,string"`
	ZljlrD20 float64 `gorm:"comment:20 日主力净流入" json:"zljlr_d20,string"`
	Zgb      string  `json:"zgb"`
	Lzg      struct {
		Code string  `json:"code"`
		Name string  `json:"name"`
		Zxj  float64 `gorm:"comment:最新价" json:"zxj,string"`
		Zdf  float64 `gorm:"comment:涨跌幅" json:"zdf,string"`
		Zd   float64 `gorm:"comment:涨幅" json:"zxd,string"`
	} `gorm:"embedded;embeddedPrefix:lzg_" json:"lzg"`
	StockType string `gorm:"comment:类型" json:"stock_type"`
}

type GetRankResp struct {
	Code int `json:"code"`
	Data struct {
		RankList []struct {
			InnerPlate
		} `json:"rank_list"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
	} `json:"data"`
	Msg string `json:"msg"`
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
		line.InnerKine = v.InnerKine
		line.Date, _ = time.Parse(datetime.LayoutDateWithLine, v.Date)
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
			InnerKine
			Date    string `json:"date"`
			Finance any    `json:"finance"`
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
	Code string
	Name string
	InnerKine
	Date time.Time `gorm:"type:date;comment:日期"`
}

type InnerKine struct {
	Open        float64 `gorm:"comment:开盘价" json:"open,string"`
	Last        float64 `gorm:"comment:最新价" json:"last,string"`
	High        float64 `gorm:"comment:最高价" json:"high,string"`
	Low         float64 `gorm:"comment:最低价" json:"low,string"`
	Volume      float64 `gorm:"comment:成交量" json:"volume,string"`
	Amount      float64 `gorm:"comment:成交额" json:"amount,string"`
	Exchange    float64 `gorm:"comment:换手率" json:"exchange,string"`
	ExchangeRaw float64 `json:"exchangeRaw,string"`
	Oi          float64 `json:"oi,string"`
	TradeDays   float64 `json:"traceDays,string"`
	Dividend    string  `json:"dividend"`
	AddZdf      string  `json:"addZdf"`
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

func getBody(url string, rsp any) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.New(err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	if err != nil {
		return errors.New(err.Error())
	}
	err = json.Unmarshal(body, rsp)
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
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
