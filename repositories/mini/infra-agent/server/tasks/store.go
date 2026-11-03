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

// ExecutionToken 表示执行器写入时使用的租约身份。
type ExecutionToken struct {
	Owner string
	Epoch int64
}

// UpdateRunStatusRequest 表示运行状态转换。
type UpdateRunStatusRequest struct {
	From []RunStatus
	To   RunStatus
	// Execution 为 nil 表示用户控制操作；执行器写入必须提供有效租约身份。
	Execution *ExecutionToken
}

// ListRunsRequest 表示运行列表的查询参数。
type ListRunsRequest struct {
	// TaskID 为空时不按任务筛选。
	TaskID string
}

// ToolAuthorizationQuery 表示工具授权的匹配条件。
type ToolAuthorizationQuery struct {
	ToolName    string
	Key         string
	ToolCallID  string
	RunID       string
	WorkspaceID string
}

// store 定义任务数据的持久化能力。
type store interface {
	CreateTask(ctx context.Context, task Task) error
	UpdateTask(ctx context.Context, taskID string, request UpdateTaskRequest) error
	CreateRun(ctx context.Context, item run) error
	// GetRun 查询运行，不存在时返回 nil、nil。
	GetRun(ctx context.Context, runID string) (*run, error)
	// UpdateRunStatus 校验原状态，并在提供 Execution 时校验租约身份及有效期，原子更新状态。
	// 更新为 cancelled 时，须在同一事务中使该运行的待审批、待输入请求失效。
	// 条件不匹配返回 false、nil；存储故障返回错误。
	UpdateRunStatus(ctx context.Context, runID string, request UpdateRunStatusRequest) (bool, error)
	// CreateRunWithAttachments 在同一事务中创建运行、更新任务的 CurrentRunID 并绑定附件。
	// 附件列表可为空；任一步失败必须回滚，提交成功后才返回 nil。
	CreateRunWithAttachments(ctx context.Context, taskID string, item run, attachmentIDs []string) error
	// ListRuns 根据查询条件获取运行列表，按创建顺序升序返回，最新运行位于末尾。
	ListRuns(ctx context.Context, request ListRunsRequest) ([]run, error)
	// GetToolExecution 从完整的执行记录中还原工具状态，无记录时返回零值。
	// 读取失败或记录损坏必须返回错误，不得视为尚未执行。
	GetToolExecution(ctx context.Context, runID, toolCallID string) (ToolExecution, error)
	// FindToolAuthorization 按工具名和匹配键查询适用于当前调用、运行或工作区的授权。
	// 未找到返回 nil、nil；空的范围标识不得匹配对应范围的授权。
	FindToolAuthorization(ctx context.Context, query ToolAuthorizationQuery) (*ToolAuthorization, error)
	// CreateToolAuthorization 保存已批准的工具授权，并校验作用域所需的标识。
	CreateToolAuthorization(ctx context.Context, authorization ToolAuthorization) error
	// GetNetworkAuthorization 查询指定运行的联网授权，未找到返回 nil、nil。
	GetNetworkAuthorization(ctx context.Context, runID string) (*NetworkAuthorization, error)
	// CreateNetworkAuthorization 保存指定运行已获准联网的授权。
	CreateNetworkAuthorization(ctx context.Context, authorization NetworkAuthorization) error
	// ListPathAuthorizations 查询工作区的路径授权，由调用方匹配路径覆盖范围及读写权限。
	ListPathAuthorizations(ctx context.Context, workspaceID string) ([]PathAuthorization, error)
	// CreatePathAuthorization 保存已批准的路径授权。
	CreatePathAuthorization(ctx context.Context, authorization PathAuthorization) error
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
