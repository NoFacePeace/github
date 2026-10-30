package tasks

import "context"

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

// runManager 管理任务运行的生命周期。
type runManager struct{}

// newRunManager 创建运行管理器。
func newRunManager() *runManager {
	return &runManager{}
}

// List 根据任务 ID 获取运行列表。
func (m *runManager) List(ctx context.Context, taskID string) ([]run, error) {
	// TODO: 根据任务 ID 从存储层查询运行列表。
	return nil, nil
}

// start 启动一次任务运行。
func (m *runManager) start(ctx context.Context, item run) error {
	return nil
}
