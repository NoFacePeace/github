package runner

import "time"

// Config 保存 Runner 的配置。
type Config struct {
	RunnerID  string
	Capacity  int
	Engines   []string
	Heartbeat HeartbeatConfig
}

// HeartbeatConfig 定义心跳上报间隔和有效期。
type HeartbeatConfig struct {
	Interval time.Duration
	TTL      time.Duration
}

// DefaultConfig 返回 Runner 的默认配置。
func DefaultConfig() Config {
	return Config{
		Heartbeat: HeartbeatConfig{
			Interval: 10 * time.Second,
			TTL:      30 * time.Second,
		},
	}
}
