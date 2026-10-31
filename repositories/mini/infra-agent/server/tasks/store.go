package tasks

import (
	"context"
	"time"
)

// UpdateTaskRequest 表示任务的部分更新参数。
// nil 表示不修改，非 nil 表示使用指针指向的值，包括空字符串。
type UpdateTaskRequest struct {
	Title        *string
	CurrentRunID *string
}

// ListRunsRequest 表示运行列表的查询参数。
type ListRunsRequest struct {
	// TaskID 为空时不按任务筛选。
	TaskID string
}

// store 定义任务数据的持久化能力。
type store interface {
	CreateTask(ctx context.Context, task Task) error
	UpdateTask(ctx context.Context, taskID string, request UpdateTaskRequest) error
	CreateRun(ctx context.Context, item run) error
	// CreateRunWithAttachments 在同一事务中创建运行、更新任务的 CurrentRunID 并绑定附件。
	// 附件列表可为空；任一步失败必须回滚，提交成功后才返回 nil。
	CreateRunWithAttachments(ctx context.Context, taskID string, item run, attachmentIDs []string) error
	// ListRuns 根据查询条件获取运行列表，按创建顺序升序返回，最新运行位于末尾。
	ListRuns(ctx context.Context, request ListRunsRequest) ([]run, error)
	// AppendEvent 追加一条业务事件。
	AppendEvent(ctx context.Context, event Event) error
	// AcquireLease 原子获取运行租约并递增版本，返回租约及是否获取成功。
	// 未获取到租约时返回 nil、false、nil，存储操作异常时返回错误。
	AcquireLease(ctx context.Context, runID, owner string, ttl time.Duration) (*lease, bool, error)
	// RenewLease 为身份匹配且尚未过期的租约续期，返回是否成功。
	RenewLease(ctx context.Context, runID, owner string, epoch int64, ttl time.Duration) (bool, error)
	// ReleaseLease 仅释放指定持有者和版本对应的租约。
	ReleaseLease(ctx context.Context, runID, owner string, epoch int64) error
	// ValidateLease 检查持有者、版本和有效期是否仍然匹配。
	ValidateLease(ctx context.Context, runID, owner string, epoch int64) (bool, error)
}
