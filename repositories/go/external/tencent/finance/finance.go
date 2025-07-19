package finance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
)

var (

	// default
	DefaultLimit = 370
)

type Point struct {
	Last   float64   `json:"last"`
	Date   time.Time `json:"date"`
	Volume float64   `json:"volume"`
}

type Stock struct {
	Code string
	Name string
}

func GetAllKline(code string, options ...Option) ([]Point, error) {
	toDate := time.Now()
	ret := []Point{}
	for {
		options = append(options, WithDate(toDate))
		sub, err := GetKline(code, options...)
		if err != nil {
			return nil, fmt.Errorf("get all kline %s error: [%w]", code, err)
		}
		ret = append(sub, ret...)
		if len(sub) < DefaultLimit {
			break
		}
		toDate = datetime.Yesterday(sub[0].Date)
	}
	return ret, nil
}

func GetKline(code string, options ...Option) ([]Point, error) {
	base := "https://proxy.finance.qq.com/cgi/cgi-bin/stockinfoquery/kline/app/get"
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("get kline %s error: [%w]", code, err)
	}
	params := url.Values{}
	params.Set("code", code)
	params.Set("limit", strconv.Itoa(DefaultLimit))
	params.Set("fqtype", BeforeAdjust.String())
	for _, option := range options {
		option.apply(&params)
	}
	u.RawQuery = params.Encode()
	var resp getKlineResp
	if err := get(u.String(), &resp); err != nil {
		return nil, fmt.Errorf("get kline %s error: [%w]", code, err)
	}
	arr := []Point{}
	for _, node := range resp.Data.Nodes {
		p := Point{}
		p.Last = node.Last
		p.Volume = node.Volume
		p.Date, _ = time.Parse(datetime.LayoutDateWithDash, node.Date)
		arr = append(arr, p)
	}
	return arr, nil
}

func getBoardRankList(code string, options ...Option) (*getBoardRankListRespData, error) {
	base := "https://proxy.finance.qq.com/cgi/cgi-bin/rank/hs/getBoardRankList"
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("get board rank list %s error: [%w]", code, err)
	}
	params := url.Values{}
	params.Set("_appver", "11.14.0")
	params.Set("offset", "0")
	params.Set("count", "40")
	params.Set("board_code", code)
	params.Set("sort_type", "marketValue")
	params.Set("direct", "down")
	for _, option := range options {
		option.apply(&params)
	}
	u.RawQuery = params.Encode()
	var resp getBoardRankListResp
	if err := get(u.String(), &resp); err != nil {
		return nil, fmt.Errorf("get board rank list %s error: [%w]", code, err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("get board rank list %s error: %v", code, resp.Msg)
	}
	return resp.Data, nil
}

type getKlineResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		StockCode string `json:"stockCode"`
		Nodes     []struct {
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
			Dividend    string  `json:"dividend" gorm:"type:varchar(255)"`
			AddZdf      string  `json:"addZdf" gorm:"type:varchar(255)"`
			Date        string  `json:"date"`
			Finance     any     `json:"finance"`
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

type getBoardRankListResp struct {
	Code int                       `json:"code"`
	Data *getBoardRankListRespData `json:"data"`
	Msg  string                    `json:"msg"`
}

type getBoardRankListRespData struct {
	RankList []struct {
		Code      string  `json:"code" gorm:"type:varchar(16);uniqueIndex:udx_stock"`
		Name      string  `json:"name" gorm:"type:varchar(32);index:idx_date"`
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
		State     string  `json:"state" gorm:"varchar(255)"`
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
		StockType string  `json:"stock_type" gorm:"type:varchar(16)"`
		Labels    []struct {
			Label int    `json:"label"`
			Value []any  `json:"value"`
			Name  string `json:"name"`
		} `json:"labels,omitempty"`
	} `json:"rank_list"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

func getRank(options ...Option) (*getRankRespData, error) {
	endpoint := "https://proxy.finance.qq.com/cgi/cgi-bin/rank/pt/getRank"
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("url parse error: [%w]", err)
	}
	params := &url.Values{}
	params.Add("count", "40")
	params.Add("offset", "0")
	params.Add("direct", "down")
	params.Add("board_type", "hy2")
	params.Add("sort_type", "PriceRatio")
	for _, option := range options {
		option.apply(params)
	}
	u.RawQuery = params.Encode()
	resp := &getRankResp{}
	if err := get(u.String(), resp); err != nil {
		return nil, fmt.Errorf("finance get error: [%w]", err)
	}
	return resp.Data, nil
}

type getRankResp struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data *getRankRespData `json:"data"`
}

type getRankRespData struct {
	RankList []struct {
		Code string `json:"code"`
		Hsl  string `json:"hsl"`
		Lb   string `json:"lb"`
		Ltsz string `json:"ltsz"`
		Lzg  struct {
			Code string `json:"code"`
			Name string `json:"name"`
			Zd   string `json:"zd"`
			Zdf  string `json:"zdf"`
			Zxj  string `json:"zxj"`
		} `json:"lzg"`
		Name      string `json:"name"`
		Speed     string `json:"speed"`
		StockType string `json:"stock_type"`
		Turnover  string `json:"turnover"`
		Volume    string `json:"volume"`
		Zd        string `json:"zd"`
		Zdf       string `json:"zdf"`
		ZdfD20    string `json:"zdf_d20"`
		ZdfD5     string `json:"zdf_d5"`
		ZdfD60    string `json:"zdf_d60"`
		ZdfW52    string `json:"zdf_w52"`
		ZdfY      string `json:"zdf_y"`
		Zgb       string `json:"zgb"`
		Zljlr     string `json:"zljlr"`
		ZljlrD20  string `json:"zljlr_d20"`
		ZljlrD5   string `json:"zljlr_d5"`
		Zllc      string `json:"zllc"`
		Zllr      string `json:"zllr"`
		Zsz       string `json:"zsz"`
		Zxj       string `json:"zxj"`
	} `json:"rank_list"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

func get(url string, resp any) error {
	rsp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get url %s error: [%w]", url, err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("url %s io read all error: [%w]", url, err)
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("url %s status code %d error: [%s]", url, rsp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, resp); err != nil {
		return fmt.Errorf("url %s json unmarshal error: [%w]", url, err)
	}
	return nil
}
