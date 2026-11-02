package tasks

import "time"

const (
	defaultTitleTimeout   = 30 * time.Second
	defaultLeaseTTL       = 30 * time.Second
	defaultCleanupTimeout = 5 * time.Second
)

// Config 表示任务服务配置。
type Config struct {
	// DefaultModel 是默认模型标识。
	DefaultModel string
	// AutoTitle 表示是否自动生成任务标题。
	AutoTitle bool
	// MaxTitleRunes 是任务标题允许的最大 Unicode 码点数，非正数表示不限制。
	MaxTitleRunes int
	// TitleTimeout 是异步生成和更新标题的超时时间，非正数时使用默认值。
	TitleTimeout time.Duration
	// LeaseTTL 是执行租约的有效期，非正数时使用默认值。
	LeaseTTL time.Duration
	// CleanupTimeout 是执行结束后释放租约的超时时间，非正数时使用默认值。
	CleanupTimeout time.Duration
}
