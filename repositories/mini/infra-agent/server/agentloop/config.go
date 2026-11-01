package agentloop

const defaultTrimTriggerPercent = 80

// Config 定义执行循环的配置。
type Config struct {
	// TrimTriggerPercent 表示达到输入预算的百分之多少时触发裁剪。
	TrimTriggerPercent int
}
