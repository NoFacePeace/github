package tasks

import "time"

// lease 表示本轮执行持有的租约凭证。
type lease struct {
	Owner     string
	Epoch     int64
	ExpiresAt time.Time
}
