package tasks

// RunStatus 表示运行的生命周期状态。
type RunStatus string

const (
	// RunStatusPending 表示等待执行。
	RunStatusPending RunStatus = "pending"
	// RunStatusRunning 表示正在执行。
	RunStatusRunning RunStatus = "running"
	// RunStatusWaitingForInput 表示等待用户输入。
	RunStatusWaitingForInput RunStatus = "waiting_for_input"
	// RunStatusWaitingForApproval 表示等待审批。
	RunStatusWaitingForApproval RunStatus = "waiting_for_approval"
	// RunStatusCompleted 表示执行完成。
	RunStatusCompleted RunStatus = "completed"
	// RunStatusFailed 表示执行失败。
	RunStatusFailed RunStatus = "failed"
	// RunStatusCancelled 表示已取消。
	RunStatusCancelled RunStatus = "cancelled"
)

// IsTerminal 判断运行状态是否已结束，空值和未知状态返回 false。
func (s RunStatus) IsTerminal() bool {
	switch s {
	case RunStatusCompleted, RunStatusFailed, RunStatusCancelled:
		return true
	default:
		return false
	}
}

// InteractionMode 表示运行的交互模式。
type InteractionMode string

const (
	// InteractionModeExecute 表示实际执行任务。
	InteractionModeExecute InteractionMode = "execute"
	// InteractionModePlan 表示只调查和规划，不执行修改。
	InteractionModePlan InteractionMode = "plan"
)

// ApprovalMode 表示运行的审批模式。
type ApprovalMode string

const (
	// ApprovalModeManual 表示需要审批的操作由人工确认。
	ApprovalModeManual ApprovalMode = "manual"
	// ApprovalModePolicy 表示根据策略决定是否需要人工确认。
	ApprovalModePolicy ApprovalMode = "policy"
	// ApprovalModeAuto 表示自动批准允许自动处理的审批。
	ApprovalModeAuto ApprovalMode = "auto"
)

// run 表示任务的一次运行。
type run struct {
	ID     string
	Status RunStatus
	// ParentRunID 为空时表示根运行。
	ParentRunID string
}
