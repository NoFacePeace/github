package tencent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NoFacePeace/github/repositories/go/utils/converter"
	"github.com/pkg/errors"
)

type VictoriaMetrics struct {
	Address string
}

func NewVictoriaMetrics(addr string) *VictoriaMetrics {
	return &VictoriaMetrics{
		Address: addr,
	}
}

func (v *VictoriaMetrics) SaveKline(kline []Kline, labels map[string]string) error {
	metric := "tencent_kline_" + labels["type"] + "_" + labels["class"]
	fields := []string{
		"open",
		"last",
		"high",
		"low",
		"volume",
		"amount",
		"exchange",
	}
	for _, field := range fields {
		vm := VictoriaMetricsData{}
		vm.Metric = map[string]string{
			"__name__": metric + "_" + field,
		}
		for _, line := range kline {
			vm.Metric["code"] = line.Code
			vm.Metric["name"] = line.Name
			switch field {
			case "open":
				vm.Values = append(vm.Values, line.Open)
			case "last":
				vm.Values = append(vm.Values, line.Last)
			case "high":
				vm.Values = append(vm.Values, line.High)
			case "low":
				vm.Values = append(vm.Values, line.Low)
			case "volume":
				vm.Values = append(vm.Values, line.Volume)
			case "amount":
				vm.Values = append(vm.Values, line.Amount)
			case "exchange":
				vm.Values = append(vm.Values, line.Exchange)
			}
			vm.Timestamps = append(vm.Timestamps, line.Date.UnixMilli())
		}
		if err := v.Write(vm); err != nil {
			return err
		}
	}
	return nil
}

func (v *VictoriaMetrics) SavePlateBoard(plates []Plate, labels map[string]string) error {
	metric := "tencent_plate_board"
	for _, label := range labels {
		metric += "_" + label
	}
	for _, plate := range plates {
		m := map[string]any{}
		if err := converter.Convert(plate, &m); err != nil {
			return err
		}
		for key, val := range m {
			var num float64
			switch val := val.(type) {
			case string:
				var err error
				num, err = strconv.ParseFloat(val, 64)
				if err != nil {
					continue
				}
			default:
				continue
			}
			vm := VictoriaMetricsData{}
			vm.Metric = map[string]string{
				"__name__": metric + "_" + key,
				"code":     plate.Code,
				"name":     plate.Name,
			}
			vm.Timestamps = append(vm.Timestamps, plate.Date.UnixMilli())
			vm.Values = append(vm.Values, num)
			if err := v.Write(vm); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VictoriaMetrics) SaveStockBoard(stocks []Stock, labels map[string]string) error {
	metric := "tencent_stock_board"
	for _, label := range labels {
		metric += "_" + label
	}
	for _, stock := range stocks {
		m := map[string]any{}
		if err := converter.Convert(stock, &m); err != nil {
			return err
		}
		for key, val := range m {
			var num float64
			switch val := val.(type) {
			case string:
				var err error
				num, err = strconv.ParseFloat(val, 64)
				if err != nil {
					continue
				}
			default:
				continue
			}
			vm := VictoriaMetricsData{}
			vm.Metric = map[string]string{
				"__name__": metric + "_" + key,
				"code":     stock.Code,
				"name":     stock.Name,
			}
			vm.Timestamps = append(vm.Timestamps, stock.Date.UnixMilli())
			vm.Values = append(vm.Values, num)
			if err := v.Write(vm); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VictoriaMetrics) Write(data VictoriaMetricsData) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.New(err.Error())
	}
	resp, err := http.Post(v.Address, "application/json", bytes.NewBuffer(bs))
	if err != nil {
		return errors.New(err.Error())
	}
	if resp.StatusCode != 204 {
		return errors.New(resp.Status)
	}
	return nil
}

type VictoriaMetricsData struct {
	Metric     map[string]string `json:"metric"`
	Values     []float64         `json:"values"`
	Timestamps []int64           `json:"timestamps"`
}
