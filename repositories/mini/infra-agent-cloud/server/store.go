package server

import (
	"context"
	"time"
)

// Store 定义任务创建流程所需的数据存取接口。
type Store interface {
	GetAgent(ctx context.Context, agentID string) (*Agent, error)
	GetVersion(ctx context.Context, versionID string) (*AgentVersion, error)
	CreateTask(ctx context.Context, task *Task) error
	CreateSession(ctx context.Context, session *Session) error
	// EnqueueJob 按 SessionID 原子去重，首次提交时保存 Job 并加入待调度队列。
	// 重复提交不重复入队，也不覆盖已有 Job；若所属 Task 不一致则返回冲突错误。
	// 数据库实现须通过 SessionID 唯一约束保证并发安全，不能只依赖先查后插。
	EnqueueJob(ctx context.Context, job *Job) error
	// ClaimJob 每次在短事务中领取一条可调度任务，无任务时返回 nil, nil。
	// 候选任务须没有领取 token 或领取已过期；按入队顺序选取。
	// 未超过 Deadline 的任务还须 NextAttemptAt 为零值或已到期，才允许领取。
	// 已超过 Deadline 的任务忽略 NextAttemptAt，允许领取后由 Service 判断超时并收尾，
	// 但仍不能抢占尚未过期的有效领取。
	// MySQL 实现通过 FOR UPDATE SKIP LOCKED 配合 LIMIT 1 锁定记录，
	// 写入新的随机领取 token、领取过期时间及处理中状态，提交后释放行锁并返回 Job。
	// 分发在事务外执行；领取未过期的任务不能被再次领取，领取过期则允许其他调度器接手。
	// 排队截止与领取过期是不同概念；Service 对超时任务标记失败，不再分发或安排重试。
	// 后续续期和状态更新须校验领取 token，防止旧持有者覆盖新持有者的结果。
	// TODO: 补充调度状态和排队时间字段，以及完成领取、失败恢复和超时处理逻辑。
	ClaimJob(ctx context.Context) (*Job, error)
	// RenewJobClaim 校验领取 token 且领取尚未过期后，按 ttl 延长占有期限。
	// ttl 必须为正数，过期时间以存储端当前时间为基准计算。
	// 任务不存在、已结束领取或 token 失效时返回 false, nil。
	RenewJobClaim(ctx context.Context, sessionID, claimToken string, ttl time.Duration) (bool, error)
	// FailJob 在同一事务中校验领取 token 和有效期，将 Job 及对应 Session 标记为失败并保存原因。
	// 同时清除领取信息，使 Job 不再参与调度；不得覆盖已结束的 Session 状态。
	// 成功返回 true, nil；领取失效或 Job 已结束调度时返回 false, nil；存储异常返回错误。
	FailJob(ctx context.Context, sessionID, claimToken, reason string) (bool, error)
	// RetryJob 原子校验领取 token 和有效期，记录失败原因并设置下次允许领取的时间。
	// 清除领取 token 和领取过期时间，使原 Job 可在 nextAttemptAt 到期后重新领取。
	// 保留原 QueuedAt、Deadline 及已确定的执行消息和投递阶段，避免延长总期限或重复分配执行。
	// 仅允许未结束调度且未超过 Deadline 的 Job 重试，不将 Session 标记为失败。
	// 成功返回 true, nil；领取失效、调度已结束或已超时返回 false, nil；存储异常返回错误。
	RetryJob(ctx context.Context, sessionID, claimToken, reason string, nextAttemptAt time.Time) (bool, error)
	// ReserveExecution 原子校验调度领取 token 和有效期，并为 Job 预留目标 Runner 的并发名额。
	// 须检查 Runner 在线、支持 Job 的引擎且仍有容量；已绑定 Runner 的任务不得改投其他节点。
	// 首次预留生成执行 token，以存储端当前时间加 ttl 设置执行租约期限，ttl 必须为正数。
	// 同时保存 Job 和 Session 的 Runner 绑定，将 Job 状态设为 dispatching，返回更新后的 Job。
	// 重复预留须复用已保存且有效的同一执行租约，不能重复占用名额或更换目标 Runner。
	// 已有执行租约失效时返回错误，由调用方处理恢复，不得静默生成新租约重新投递。
	// 领取失效、Runner 不可用、容量不足或存储异常均返回错误；本方法不投递消息或启动执行。
	ReserveExecution(ctx context.Context, sessionID, claimToken, runnerID string, ttl time.Duration) (*Job, error)
	// ListRunners 返回执行节点的状态、引擎能力和容量快照，无节点时返回空列表。
	// Service 用于查找已绑定节点或选择候选节点，实际预留仍须原子复查节点状态与容量。
	ListRunners(ctx context.Context) ([]Runner, error)
	// PublishJob 将已确定的执行消息幂等投递至 job.RunnerID 对应的执行通道。
	// 须校验执行租约有效，使用固定的 MessageID 和消息内容；重试不得重复发布同一消息。
	// 发布去重标记与消息写入须原子完成；不向 Runner 暴露调度领取 token 等内部字段。
	// 成功仅表示消息已投递，不表示 Runner 已开始执行；结果不确定时返回错误，供同消息重试。
	PublishJob(ctx context.Context, job *Job) error
	// CompleteJobDispatch 原子校验调度领取 token 和有效期，将 Job 标记为 dispatched 并清除领取信息。
	// 保留执行 token 和执行租约期限，Session 保持 pending，等待 Runner 上报开始执行。
	// 同一有效领取下重复确认已投递记录应视为成功，不覆盖取消或失败等终态。
	// 成功返回 true, nil；领取失效或状态不允许确认时返回 false, nil；存储异常返回错误。
	CompleteJobDispatch(ctx context.Context, sessionID, claimToken string) (bool, error)
}
