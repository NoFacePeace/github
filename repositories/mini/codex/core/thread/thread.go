package thread

import "context"

// TurnInput 定义轮次输入，具体消息与附加上下文字段后续补充。
type TurnInput struct{}

// TurnSubmission 表示输入已被接收；Started 为真表示新建轮次，否则表示追加到已有轮次。
type TurnSubmission struct {
	TurnID  string
	Started bool
}

// Operation 定义控制操作，后续补充中断、手动压缩和审批响应等操作协议。
type Operation struct{}

// Event 定义执行期间的事件，后续补充轮次标识、文本增量、工具调用和结束状态等字段。
type Event struct{}

// Status 定义对话运行状态快照，具体状态与活动轮次字段后续补充。
type Status struct{}

// TokenUsage 定义输入、输出和累计用量，具体统计字段后续补充。
type TokenUsage struct{}

// HistoryItem 定义一条持久化历史记录，具体消息与执行记录字段后续补充。
type HistoryItem struct{}

// Settings 定义可更新的对话配置，具体模型与权限等字段后续补充。
type Settings struct{}

// Thread 定义单个对话的公开操作接口，封装内部会话运行时和输入输出通道。
type Thread interface {
	// ID 返回对话的稳定标识，恢复对话时保持不变。
	ID() string

	// StartTurn 仅在空闲时接收输入并启动新轮次，返回轮次标识，不等待执行完成。
	// 对话忙碌或已关闭时返回错误；上下文仅控制提交过程，停止执行须提交中断操作。
	StartTurn(ctx context.Context, input TurnInput) (string, error)

	// StartOrSteerTurn 在空闲时启动新轮次，在普通轮次运行时追加输入。
	// 返回实际接收输入的轮次标识与接收方式；无法接收时返回错误。
	StartOrSteerTurn(ctx context.Context, input TurnInput) (TurnSubmission, error)

	// SteerTurn 仅向指定的活动普通轮次追加输入，轮次已结束或不支持追加时返回错误。
	SteerTurn(ctx context.Context, turnID string, input TurnInput) error

	// NextEvent 等待并消费下一个执行事件，调用方应使用单个读取者持续接收。
	// 上下文取消只结束本次等待；事件流关闭且已消费完毕时返回 io.EOF。
	NextEvent(ctx context.Context) (Event, error)

	// Submit 接收控制操作并返回操作标识，执行结果通过事件流报告。
	// 中断只停止当前轮次并保留对话；手动压缩作为独立任务执行。
	Submit(ctx context.Context, operation Operation) (string, error)

	// Status 查询对话当前的运行状态和活动轮次快照。
	Status(ctx context.Context) (Status, error)

	// TokenUsage 查询最近一轮与对话累计的模型用量，首次统计前可以返回空值。
	TokenUsage(ctx context.Context) (*TokenUsage, error)

	// LoadHistory 读取对话的持久化历史，不启动或恢复轮次。
	LoadHistory(ctx context.Context) ([]HistoryItem, error)

	// UpdateSettings 校验并更新对话配置，等待接受或拒绝，不修改已经发出的模型请求。
	UpdateSettings(ctx context.Context, settings Settings) error

	// Close 发起对话关闭，停止当前任务、保存历史并等待运行时退出。
	// 重复调用应保持安全；上下文取消只停止等待，已经发起的关闭继续进行。
	Close(ctx context.Context) error
}
