package gu

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
)

func HKMinute(code string) ([]Point, error) {
	var resp HKMinuteResp
	url := "https://web.ifzq.gtimg.cn/appstock/app/hkMinute/query?code=" + code
	if err := get(url, &resp); err != nil {
		return nil, fmt.Errorf("get error: [%w]", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("response code: %d, message: %s", resp.Code, resp.Msg)
	}
	strs := resp.Data[code].Data.Data
	date := resp.Data[code].Data.Date
	points := []Point{}
	turnover := 0.0
	volume := 0.0
	for _, str := range strs {
		arr := strings.Split(str, " ")
		price, _ := strconv.ParseFloat(arr[1], 64)
		v, _ := strconv.ParseFloat(arr[2], 64)
		dt, _ := time.Parse(datetime.LayoutDatetimeMinue, date+arr[0])
		turnover += price * (v - volume)
		volume = v
		avg := turnover / volume
		points = append(points, Point{
			Price: price,
			Date:  dt,
			AVG:   avg,
			Dist:  (price/avg - 1) * 100,
		})
	}
	return points, nil
}

type Point struct {
	Price float64
	Date  time.Time
	AVG   float64
	Dist  float64
}

type HKMinuteResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data map[string]struct {
		Data struct {
			Data []string `json:"data"`
			Date string   `json:"date"`
		} `json:"data"`
		Qt struct {
			Hk00700 []string `json:"hk00700"`
			Market  []string `json:"market"`
		} `json:"qt"`
		Vcm string `json:"vcm"`
	} `json:"data"`
}

func get(url string, resp any) error {
	rsp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get %s error: [%w]", url, err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("io read all error: [%w]", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("get %s error: status code %d, body %s", url, rsp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, resp); err != nil {
		return fmt.Errorf("json unmarshal error: [%w]", err)
	}
	return nil
}
