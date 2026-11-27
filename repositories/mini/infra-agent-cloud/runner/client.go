package runner

import (
	"context"
	"time"
)

// Client 封装 Runner 对 Server 的 HTTP 调用。
type Client interface {
	// RenewExecution 向 Server 请求执行续约，返回确认的过期时间；须区分授权拒绝与临时通信错误。
	RenewExecution(ctx context.Context, job *Job) (time.Time, error)
}
