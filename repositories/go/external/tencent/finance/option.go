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

func WithCount(count int) Option {
	return &countOption{
		count: count,
	}
}

func WithOffset(offset int) Option {
	return &offsetOption{
		offset: offset,
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

type countOption struct {
	count int
}

func (o *countOption) apply(params *url.Values) {
	params.Set("count", strconv.Itoa(o.count))
}

type offsetOption struct {
	offset int
}

func (o *offsetOption) apply(params *url.Values) {
	params.Set("offset", strconv.Itoa(o.offset))
}

type limitOption int

func (l limitOption) apply(params *url.Values) {
	params.Set("limit", strconv.Itoa(int(l)))
}

func WithLimit(limit int) Option {
	return limitOption(limit)
}

type directOption string

func (d directOption) apply(params *url.Values) {
	params.Set("direct", string(d))
}

var (
	DirectOptionUp   directOption = "up"
	DirectOptionDown directOption = "down"
)
