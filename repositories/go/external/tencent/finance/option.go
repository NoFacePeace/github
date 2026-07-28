package finance

import (
	"net/url"
	"strconv"
	"time"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
)

type Option func(*url.Values)

type AdjustType string

var (
	BeforeAdjust AdjustType = "qfq"
	AfterAdjust  AdjustType = "hfq"
	NoneAdjust   AdjustType = ""
)

func WithAdjuct(ad AdjustType) Option {
	return func(params *url.Values) {
		params.Set("fqtype", ad.String())
	}
}

func WithCount(count int) Option {
	return func(params *url.Values) {
		params.Set("count", strconv.Itoa(count))
	}
}

func WithOffset(offset int) Option {
	return func(params *url.Values) {
		params.Set("offset", strconv.Itoa(offset))
	}
}

func (ad AdjustType) String() string {
	return string(ad)
}

func WithDate(date time.Time) Option {
	return func(params *url.Values) {
		params.Set("toDate", date.Format(datetime.LayoutDateWithDash))
	}
}

func WithLimit(limit int) Option {
	return func(params *url.Values) {
		params.Set("limit", strconv.Itoa(limit))
	}
}

var (
	DirectOptionUp Option = func(params *url.Values) {
		params.Set("direct", "up")
	}
	DirectOptionDown Option = func(params *url.Values) {
		params.Set("direct", "down")
	}
)
