package task

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/NoFacePeace/github/repositories/go/external/tencent/finance"
	"github.com/NoFacePeace/github/repositories/go/projects/cron/channel"
	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
)

type Config struct {
	Name     string
	Spec     string
	Channels []ChannelConfig
}

type ChannelConfig struct {
	Name   string
	Config map[string]any
}

type Task struct {
	name       string
	channels   []channel.Channel
	spec       string
	fn         func() Result
	lastResult *Result
}

type Result struct {
	Time    time.Time
	Level   slog.Level
	Message string
}

func (r *Result) String() string {
	return fmt.Sprintf("time: %s\nlevel: %s\nmessage\n%s", r.Time, r.Level, r.Message)
}

func New(name, spec string, channels []channel.Channel, fn func() Result) *Task {
	return &Task{
		name:     name,
		spec:     spec,
		channels: channels,
		fn:       fn,
	}
}

func (t *Task) Spec() string {
	return t.spec
}

func (t *Task) Start() Result {
	return t.fn()
}

func (t *Task) Channels() []channel.Channel {
	return t.channels
}

func GetPlatesDroppedBy20Percent() Result {
	ret := Result{}
	ret.Time = time.Now()
	if !datetime.IsChinesStockMarketTradingDay(ret.Time) {
		ret.Message = "today is weekday or outside trading hours"
		return ret
	}
	plates, err := finance.ListPlates(finance.PlateTypeHY2)
	if err != nil {
		ret.Message = "finance list plates error"
		slog.Error("finance list plates error", "error", err)
		ret.Level = slog.LevelError
		return ret
	}
	m := []map[string]any{}
	for _, plate := range plates {
		ps, err := finance.GetKline(plate.Code)
		if err != nil {
			slog.Error("finance get kline error", "error", err)
			ret.Level = slog.LevelError
			ret.Message = "finance get kline error"
			return ret
		}
		if len(ps) == 0 {
			continue
		}
		n := len(ps)
		last := ps[n-1]
		mn := last.Last
		mx := last.Last
		date := last.Date
		for i := n - 1; i >= 0; i-- {
			if ps[i].Last < mn {
				break
			}
			if ps[i].Last >= mx {
				mx = ps[i].Last
				date = ps[i].Date
			}
		}
		percent := (mn - mx) / mx * 100
		if percent == 0 {
			break
		}
		if percent >= -20 {
			continue
		}
		tmp := map[string]any{
			"name":    plate.Name,
			"percent": percent,
			"date":    date,
		}
		m = append(m, tmp)
	}
	msg := ""
	for _, v := range m {
		msg += fmt.Sprintf("name: %v, date: %v, percent: %v\n", v["name"], v["date"], v["percent"])
	}
	ret.Message = msg
	ret.Level = slog.LevelInfo
	return ret
}
