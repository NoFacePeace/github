// Package tasks 提供任务相关的领域模型和业务逻辑。
package tasks

// TriggerType 表示任务的触发类型。
type TriggerType string

const (
	// TriggerTypeManual 表示手动触发。
	TriggerTypeManual TriggerType = "manual"
	// TriggerTypeSchedule 表示定时触发。
	TriggerTypeSchedule TriggerType = "schedule"
	// TriggerTypeWebhook 表示通过 Webhook 触发。
	TriggerTypeWebhook TriggerType = "webhook"
	// TriggerTypeIM 表示通过即时消息触发。
	TriggerTypeIM TriggerType = "im"
	// TriggerTypeEvent 表示通过事件触发。
	TriggerTypeEvent TriggerType = "event"
)

// Task 表示一个任务。
type Task struct {
	TriggerType TriggerType
}

// createRun 创建任务的一次运行。
func (t Task) createRun() run {
	return run{}
}
