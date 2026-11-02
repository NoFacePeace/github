package agentloop

const (
	defaultTrimTriggerPercent   = 80
	defaultTrimRetentionPercent = 50
)

// Config 定义执行循环的配置。
type Config struct {
	// TrimTriggerPercent 表示达到输入预算的百分之多少时触发裁剪。
	TrimTriggerPercent int
	// TrimRetentionPercent 表示删除消息块后希望保留的字节比例，基于去重和截短后的大小计算。
	TrimRetentionPercent int
}
