package thread

import "context"

// config 保存创建、恢复或分叉对话所需的内部配置，具体字段后续补充。
type config struct{}

// Option 定义单个可选配置函数，按传入顺序应用到默认配置，相同字段以后传入的选项为准。
type Option func(*config)

// Manager 定义多个对话的管理接口，负责创建、查找、恢复和关闭对话运行实例。
type Manager interface {
	// StartThread 创建新的对话，初始化内部运行时并注册到管理器；未传入选项时使用默认配置。
	StartThread(ctx context.Context, options ...Option) (Thread, error)

	// GetThread 获取已加载的对话；不存在时返回错误，不自动恢复历史。
	GetThread(ctx context.Context, threadID string) (Thread, error)

	// ListThreadIDs 列出当前管理器中已加载的对话标识，不查询全部持久化历史。
	ListThreadIDs(ctx context.Context) ([]string, error)

	// ResumeThread 根据持久化历史恢复指定对话，沿用原对话标识并注册运行实例。
	ResumeThread(ctx context.Context, threadID string, options ...Option) (Thread, error)

	// ForkThread 基于指定对话的历史创建具有新标识的独立对话。
	ForkThread(ctx context.Context, threadID string, options ...Option) (Thread, error)

	// RemoveThread 从管理器移除并返回对话句柄；不关闭运行实例，也不删除历史。
	RemoveThread(ctx context.Context, threadID string) (Thread, error)

	// ShutdownAllThreads 关闭已加载的对话并等待退出，等待时间由上下文限制。
	ShutdownAllThreads(ctx context.Context) error

	// SubscribeThreadCreated 订阅注册新对话的通知；上下文取消后结束订阅并关闭通道。
	SubscribeThreadCreated(ctx context.Context) (<-chan string, error)
}
