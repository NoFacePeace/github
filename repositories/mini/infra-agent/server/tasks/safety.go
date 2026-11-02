package tasks

import (
	"context"

	"infra-agent/server/tools"
)

// SafetyEvaluator 根据安全策略评估本次工具调用。
type SafetyEvaluator interface {
	Evaluate(ctx context.Context, request SafetyRequest) (Decision, error)
}

// SafetyRequest 表示策略判断所需的工具、风险和当前执行环境。
type SafetyRequest struct {
	ToolName string
	Risk     tools.Risk
	Env      tools.Env
}

// DecisionAction 表示安全策略的决定。
type DecisionAction string

const (
	DecisionAllow           DecisionAction = "allow"
	DecisionDeny            DecisionAction = "deny"
	DecisionRequireApproval DecisionAction = "require_approval"
)

// Decision 表示安全策略的决定和原因。
type Decision struct {
	Action DecisionAction
	Reason string
}
