package tencent

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/NoFacePeace/github/repositories/go/util/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	// board type
	BoardType1 = "hy1"
	BoardType2 = "hy2"
	// board code
	BoardCode = "aStock"
)

type Tencent struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Tencent {
	return &Tencent{
		db: db,
	}
}

func (t *Tencent) GetBoard(boardType string) error {
	arr, err := getRank(boardType)
	if err != nil {
		return err
	}
	// date, _ := strconv.Atoi(time.Now().Local().Format("20060102"))

	for _, v := range arr {
		v.Md5, err = crypto.Md5Sum(v)
		if err != nil {
			return err
		}
		v.Date = time.Now().Local()
		ret := t.db.Select("md5").Where("name = ?", v.Name).Where("md5 = ?", v.Md5).First(&Board{})
		if ret.Error == nil {
			continue
		}
		slog.Error(ret.Error.Error())
		tx := t.db.Create(&v)
		if tx.Error != nil {
			return errors.New(tx.Error.Error())
		}
		fmt.Println(v)
	}
	return nil
}

type Board struct {
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
	StockType string `gorm:"comment:类型"`
	Md5       string
	Date      time.Time
}

func getRank(boardType string) ([]Board, error) {
	url := "https://proxy.finance.qq.com/cgi/cgi-bin/rank/pt/getRank?direct=down&sort_type=PriceRatio&board_type=" + boardType
	var resp GetRankResp
	if err := getBody(url, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Msg)
	}
	arr := []Board{}
	for _, v := range resp.Data.RankList {
		b := Board{}
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

func getBody(url string, rsp any) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.New(err.Error())
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
