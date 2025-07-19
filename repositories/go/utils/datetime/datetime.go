package datetime

import (
	"fmt"
	"time"
)

var (
	LayoutDateWithLine  = "2006-01-02"
	LayoutDate          = "20060102"
	LayoutDateWithDash  = "2006-01-02"
	LayoutDatetimeMinue = "200601021504"
)

func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

func EqualDate(t1, t2 time.Time) bool {
	y1, y2 := t1.Year(), t2.Year()
	m1, m2 := t1.Month(), t2.Month()
	d1, d2 := t1.Day(), t2.Day()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func Yesterday(ts ...time.Time) time.Time {
	if len(ts) != 0 {
		return ts[0].AddDate(0, 0, -1)
	}
	return time.Now().AddDate(0, 0, -1)
}

func IsChinesStockMarketTradingDay(t time.Time) bool {
	if IsWeekend(t) {
		return false
	}
	year, month, day := t.Date()
	loc := t.Location()
	start := time.Date(year, month, day, 9, 30, 0, 0, loc)
	end := time.Date(year, month, day, 15, 0, 0, 0, loc)
	fmt.Println(start, end)
	return t.After(start) && t.Before(end)
}
