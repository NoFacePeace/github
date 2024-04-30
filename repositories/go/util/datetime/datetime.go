package datetime

import "time"

var (
	LayoutDateWithLine = "2006-01-02"
)

func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}
