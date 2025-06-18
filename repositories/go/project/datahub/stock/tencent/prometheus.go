package tencent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/castai/promwrite"
)

type Prometheus struct {
	Client *promwrite.Client
}

type Config struct {
	Address string
}

var (
	MetricAstockPlateBoardZxj = "astock_plate_board_zxj"
	MetricAstockPlateKlineDay = "tencent_astock_plate_kline_day"
)

func NewPrometheus(c *Config) *Prometheus {
	client := promwrite.NewClient(c.Address)
	return &Prometheus{
		Client: client,
	}
}

func (p *Prometheus) History() {
	slog.Info("start to scape history")
	p.HistoryAstockPlate()
	slog.Info("end to scape history")
}

func (p *Prometheus) HistoryAstockPlate() {
	// 行业2
	plates, err := getFullRank(BoardType2)
	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return
	}
	for _, v := range plates {
		slog.Info(fmt.Sprintf("start to scrape history kline: %v", v.Name))
		slog.Info(fmt.Sprintf("end to scrape history kline: %v", v.Name))
	}
}

func (p *Prometheus) Write(ps []Point) error {
	arr := []promwrite.TimeSeries{}
	for _, v := range ps {
		ts := promwrite.TimeSeries{}
		ts.Labels = append(ts.Labels, promwrite.Label{
			Name:  "__name__",
			Value: v.Metric,
		})
		for k, v := range v.Labels {
			ts.Labels = append(ts.Labels, promwrite.Label{
				Name:  k,
				Value: v,
			})
		}
		ts.Sample = promwrite.Sample{
			Time:  v.Time,
			Value: v.Value,
		}
		arr = append(arr, ts)
	}
	_, err := p.Client.Write(context.Background(), &promwrite.WriteRequest{
		TimeSeries: arr,
	})
	return err
}

type Point struct {
	Metric string
	Labels map[string]string
	Time   time.Time
	Value  float64
}
