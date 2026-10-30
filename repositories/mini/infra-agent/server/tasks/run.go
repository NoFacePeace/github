package tasks

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
type run struct{}
