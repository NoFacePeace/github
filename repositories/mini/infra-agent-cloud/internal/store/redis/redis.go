package redis

import (
	"crypto/rand"
	"net/url"

	"github.com/redis/go-redis/v9"
)

// Redis 封装 Runner 执行链路的 Redis 操作。待实现的方法如下。
//
//   - ConsumeCommands：消费发给当前 Runner 的控制命令。
//   - RemoveHeartbeat：清理当前 Runner 的心跳和在线索引。
//   - GetMaintenance：读取当前 Runner 的维护状态。
//   - GetWorkspaceCleanupResult：查询工作区清理命令的结果。
//   - CompleteWorkspaceCleanup：保存工作区清理命令的结果。
type Redis struct {
	client       *redis.Client
	jobStreamKey string
	jobGroup     string
	jobConsumer  string
}

func NewRedis(client *redis.Client, runnerID string) Redis {
	return Redis{
		client:       client,
		jobStreamKey: runnerJobsKeyPrefix + url.PathEscape(runnerID),
		jobGroup:     runnerJobsGroupPrefix + runnerID,
		jobConsumer:  runnerID + "-" + rand.Text(),
	}
}
