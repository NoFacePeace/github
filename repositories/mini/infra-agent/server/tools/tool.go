// Package tools 提供工具的定义与实现。
package tools

import (
	"context"
	"encoding/json"
)

// Tool 定义工具的能力。
type Tool interface {
	Name() string
	// Schema 返回工具调用参数的 JSON Schema。
	Schema() json.RawMessage
	// Risk 评估本次工具调用的风险。
	Risk(ctx context.Context, input json.RawMessage, env Env) (Risk, error)
	// ApprovalRequirement 返回工具自身的审批要求。
	ApprovalRequirement() ApprovalRequirement
	// AuthorizationKey 返回本次调用的工具授权匹配键，不包含联网和路径权限。
	// 同一工具中返回相同键的调用可以共享符合范围的工具授权。
	AuthorizationKey(input json.RawMessage) (string, error)
	// Execute 执行已获授权的工具调用，返回工具结果。
	Execute(ctx context.Context, input json.RawMessage, env Env) (Result, error)
}

// ApprovalRequirement 表示工具声明的审批要求，仍受运行审批模式和安全策略约束。
type ApprovalRequirement string

const (
	// ApprovalNone 表示工具自身不要求审批，仍需检查安全策略。
	ApprovalNone ApprovalRequirement = "none"
	// ApprovalRequired 表示必须获得授权，可以使用符合范围的已有授权。
	ApprovalRequired ApprovalRequirement = "required"
	// ApprovalPolicy 表示根据本次调用风险和安全策略决定是否需要审批。
	ApprovalPolicy ApprovalRequirement = "policy"
)

// Env 表示工具调用的执行环境。
type Env struct{}

// PathOperation 表示路径访问所需的操作权限。
type PathOperation string

const (
	PathRead  PathOperation = "read"
	PathWrite PathOperation = "write"
)

// PathAccess 表示本次调用需要的文件或目录访问权限。
type PathAccess struct {
	Path      string
	Operation PathOperation
}

// Risk 描述本次工具调用的风险信息。
type Risk struct {
	// RequiresNetwork 表示本次调用需要联网，是否需要审批由执行器结合已有授权判断。
	RequiresNetwork bool
	// Paths 表示本次调用需要访问的路径及操作权限。
	Paths []PathAccess
}

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
