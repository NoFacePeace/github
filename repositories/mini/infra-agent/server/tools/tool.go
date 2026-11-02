// Package tools 提供工具的定义与实现。
package tools

// Tool 定义工具的能力。
type Tool interface{}

const (
	NameAskUser      = "ask_user"
	NameGetSkillFile = "get_skill_file"
	NameUpdatePlan   = "update_plan"
)

// Status 表示工具调用的处理状态。
type Status string

const (
	StatusCompleted          Status = "completed"
	StatusWaitingForApproval Status = "waiting_for_approval"
	StatusWaitingForInput    Status = "waiting_for_input"
)

// Result 表示一次工具调用的处理结果。
type Result struct {
	ToolCallID string
	Status     Status
	Content    string
}
