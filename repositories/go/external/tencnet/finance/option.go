package finance

import (
	"net/url"
	"strconv"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
)

type Option interface {
	apply(*url.Values)
}

type AdjustType string

var (
	BeforeAdjust AdjustType = "qfq"
	AfterAdjust  AdjustType = "hfq"
	NoneAdjust   AdjustType = ""
)

func WithAdjuct(ad AdjustType) Option {
	return &adjuctOption{
		option: ad,
	}
}

func (ad AdjustType) String() string {
	return string(ad)
}

type adjuctOption struct {
	option AdjustType
}

func (o *adjuctOption) apply(params *url.Values) {
	params.Set("fqtype", o.option.String())
}

func WithDate(date time.Time) Option {
	return &dateOption{
		date: date,
	}
}

type dateOption struct {
	date time.Time
}

func (o *dateOption) apply(params *url.Values) {
	params.Set("toDate", o.date.Format(datetime.LayoutDateWithDash))
}

func WithCount(count int) Option {
	return &countOption{
		count: count,
	}
}

type countOption struct {
	count int
}

func (o *countOption) apply(params *url.Values) {
	params.Set("count", strconv.Itoa(o.count))
}
