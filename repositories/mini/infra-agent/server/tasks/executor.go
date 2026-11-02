package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"infra-agent/server/agentloop"
	"infra-agent/server/llm"
	"infra-agent/server/tools"
)

// toolExecutor 协调本次执行的工具恢复、审批、执行及结果保存。
type toolExecutor struct {
	approvalMode    ApprovalMode
	safetyEvaluator SafetyEvaluator
	recorder        agentloop.Recorder
	tools           map[string]tools.Tool
	store           store
	taskID          string
	runID           string
	workspaceID     string
	lease           *lease
}

// Execute 批量处理工具调用，返回执行结果或等待状态。
func (e *toolExecutor) Execute(ctx context.Context, calls []llm.ToolCall) ([]tools.Result, error) {
	// 按调用顺序处理，遇到需要审批或输入的调用时暂停，后续调用保留待恢复处理。
	var results []tools.Result
	for _, call := range calls {

		// 1. 公共检查与信息准备。
		// 检查取消。
		if err := ctx.Err(); err != nil {
			return results, err
		}
		// 按名称查找工具：不存在则保存未知工具错误并返回给模型，不进入审批和执行。
		selected, exists := e.tools[call.Name]
		if !exists {
			result := tools.Result{
				ToolCallID: call.ID,
				Status:     tools.StatusCompleted,
				Content:    fmt.Sprintf("unknown tool: %s", call.Name),
			}
			if err := e.recorder.Record(ctx, agentloop.Entry{
				Type: agentloop.EntryTypeToolResult,
				Data: map[string]any{
					"tool_call_id": result.ToolCallID,
					"tool":         call.Name,
					"status":       result.Status,
					"content":      result.Content,
				},
			}); err != nil {
				return results, fmt.Errorf("record tool result for call %s: %w", call.ID, err)
			}
			results = append(results, result)
			continue
		}
		// 检查执行记录。
		record, err := e.store.GetToolExecution(ctx, e.runID, call.ID)
		if err != nil {
			return results, fmt.Errorf("get tool execution for call %s: %w", call.ID, err)
		}
		if record.Result != nil {
			results = append(results, *record.Result)
			continue
		}
		if record.Started {
			result := tools.Result{
				ToolCallID: call.ID,
				Status:     tools.StatusCompleted,
				Content:    "tool execution state is unknown after recovery; the tool was not replayed",
			}
			if err := e.recorder.Record(ctx, agentloop.Entry{
				Type: agentloop.EntryTypeToolResult,
				Data: map[string]any{
					"tool_call_id": result.ToolCallID,
					"tool":         call.Name,
					"status":       result.Status,
					"content":      result.Content,
				},
			}); err != nil {
				return results, fmt.Errorf("record unknown tool execution for call %s: %w", call.ID, err)
			}
			results = append(results, result)
			continue
		}
		// 检查参数 JSON 格式：暂不自动修复，无效则返回工具错误，让模型重新生成调用。
		if !json.Valid(call.Arguments) {
			result := tools.Result{
				ToolCallID: call.ID,
				Status:     tools.StatusCompleted,
				Content:    "tool call arguments must be valid JSON",
			}
			if err := e.recorder.Record(ctx, agentloop.Entry{
				Type: agentloop.EntryTypeToolResult,
				Data: map[string]any{
					"tool_call_id": result.ToolCallID,
					"tool":         call.Name,
					"status":       result.Status,
					"content":      result.Content,
				},
			}); err != nil {
				return results, fmt.Errorf("record invalid tool arguments for call %s: %w", call.ID, err)
			}
			results = append(results, result)
			continue
		}
		// 按工具的参数 Schema 校验必填字段、类型和约束；失败则返回工具错误，不进入审批和执行。
		if err := validateToolArguments(call.Arguments, selected.Schema()); err != nil {
			result := tools.Result{
				ToolCallID: call.ID,
				Status:     tools.StatusCompleted,
				Content:    fmt.Sprintf("tool call arguments are invalid: %v", err),
			}
			if err := e.recorder.Record(ctx, agentloop.Entry{
				Type: agentloop.EntryTypeToolResult,
				Data: map[string]any{
					"tool_call_id": result.ToolCallID,
					"tool":         call.Name,
					"status":       result.Status,
					"content":      result.Content,
				},
			}); err != nil {
				return results, fmt.Errorf("record invalid tool arguments for call %s: %w", call.ID, err)
			}
			results = append(results, result)
			continue
		}
		// Hook 暂时跳过，连续重复调用限制后续按需增加。
		// 评估风险，确定路径、副作用和所需权限。
		// TODO: 根据本次执行信息组装环境。
		env := tools.Env{}
		risk, err := selected.Risk(ctx, call.Arguments, env)
		if err != nil {
			return results, fmt.Errorf("assess tool risk for call %s: %w", call.ID, err)
		}
		// 公共检查可以阻止无效调用，但不直接决定是否请求审批。

		// 2. 按工具审批声明分支，结合审批模式和授权作出决定。
		// None：工具不主动要求审批，仍检查安全策略，再结合审批模式决定 allow、deny 或 require_approval。
		// Required：检查审批模式是否允许自动处理；否则匹配单次批准和可复用授权，未满足则 require_approval。
		// Policy：根据风险评估安全策略；对 require_approval 再结合审批模式和已有授权处理。
		// Manual 要求人工决定；Policy 按策略决定；Auto 只自动批准允许自动处理的权限。
		// 各分支统一遵守硬性限制和当前调用的拒绝决定，不被自动审批或复用授权覆盖。
		// 复用授权需匹配工具、参数范围、作用域和有效期，全部所需权限满足后才允许执行。
		var decision Decision
		switch requirement := selected.ApprovalRequirement(); requirement {
		case tools.ApprovalNone, tools.ApprovalPolicy:
			decision, err = e.safetyEvaluator.Evaluate(ctx, SafetyRequest{
				ToolName: call.Name,
				Risk:     risk,
				Env:      env,
			})
			if err != nil {
				return results, fmt.Errorf("evaluate safety for call %s: %w", call.ID, err)
			}
		case tools.ApprovalRequired:
			decision = Decision{Action: DecisionRequireApproval, Reason: "tool requires approval"}
		default:
			return results, fmt.Errorf("unknown approval requirement %q for tool %s", requirement, call.Name)
		}

		// 3. 统一处理决定。
		if err := ctx.Err(); err != nil {
			return results, err
		}
		result := tools.Result{ToolCallID: call.ID, Status: tools.StatusCompleted}
		switch decision.Action {
		case DecisionDeny:
			result.Content = "tool call denied: " + decision.Reason
		case DecisionRequireApproval:
			// 自动审批只处理需要审批的决定，不覆盖策略拒绝。
			autoApproved := false
			switch e.approvalMode {
			case "", ApprovalModeManual, ApprovalModePolicy:
				// 空值默认保留人工审批；策略模式保留安全策略的决定。
			case ApprovalModeAuto:
				autoApproved = true
			default:
				return results, fmt.Errorf("unknown approval mode %q", e.approvalMode)
			}
			// 工具、网络和路径权限分别满足后才能放行，工具授权不能替代资源授权。
			toolApproved := autoApproved
			if !toolApproved {
				key, err := selected.AuthorizationKey(call.Arguments)
				if err != nil {
					return results, fmt.Errorf("get authorization key for call %s: %w", call.ID, err)
				}
				authorization, err := e.store.FindToolAuthorization(ctx, ToolAuthorizationQuery{
					ToolName: call.Name, Key: key, ToolCallID: call.ID,
					RunID: e.runID, WorkspaceID: e.workspaceID,
				})
				if err != nil {
					return results, fmt.Errorf("find tool authorization for call %s: %w", call.ID, err)
				}
				toolApproved = authorization != nil
			}
			networkApproved := !risk.RequiresNetwork
			if !networkApproved {
				authorization, err := e.store.GetNetworkAuthorization(ctx, e.runID)
				if err != nil {
					return results, fmt.Errorf("get network authorization for call %s: %w", call.ID, err)
				}
				networkApproved = authorization != nil && e.runID != "" && authorization.RunID == e.runID
			}
			pathsApproved := autoApproved || len(risk.Paths) == 0
			if !pathsApproved && e.workspaceID != "" {
				authorizations, err := e.store.ListPathAuthorizations(ctx, e.workspaceID)
				if err != nil {
					return results, fmt.Errorf("list path authorizations for call %s: %w", call.ID, err)
				}
				pathsApproved = true
				for _, access := range risk.Paths {
					matched := false
					for _, authorization := range authorizations {
						if authorization.WorkspaceID != e.workspaceID || !filepath.IsAbs(access.Path) || !filepath.IsAbs(authorization.Path) {
							continue
						}
						relative, err := filepath.Rel(authorization.Path, access.Path)
						if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
							continue
						}
						if (access.Operation == tools.PathRead && authorization.Read) || (access.Operation == tools.PathWrite && authorization.Write) {
							matched = true
							break
						}
					}
					if !matched {
						pathsApproved = false
						break
					}
				}
			}
			if toolApproved && networkApproved && pathsApproved {
				decision.Action = DecisionAllow
			}
			if decision.Action == DecisionRequireApproval {
				if err := e.recorder.Record(ctx, agentloop.Entry{
					Type: agentloop.EntryTypeApprovalRequested,
					Data: map[string]any{
						"tool_call_id": call.ID,
						"tool":         call.Name,
						"input":        call.Arguments,
						"reason":       decision.Reason,
					},
				}); err != nil {
					return results, fmt.Errorf("record approval request for call %s: %w", call.ID, err)
				}
				results = append(results, tools.Result{
					ToolCallID: call.ID, Status: tools.StatusWaitingForApproval,
				})
				return results, nil
			}
			// 自动批准后进入统一的工具执行流程。
			fallthrough
		case DecisionAllow:
			valid, err := e.store.ValidateLease(ctx, e.runID, e.lease.Owner, e.lease.Epoch)
			if err != nil {
				return results, fmt.Errorf("validate lease before call %s: %w", call.ID, err)
			}
			if !valid {
				return results, fmt.Errorf("execution lease lost before call %s", call.ID)
			}
			if err := e.recorder.Record(ctx, agentloop.Entry{
				Type: agentloop.EntryTypeToolStarted,
				Data: map[string]any{
					"tool_call_id": call.ID,
					"tool":         call.Name,
					"input":        call.Arguments,
				},
			}); err != nil {
				return results, fmt.Errorf("record tool start for call %s: %w", call.ID, err)
			}
			if err := ctx.Err(); err != nil {
				return results, err
			}
			output, err := selected.Execute(ctx, call.Arguments, env)
			if ctx.Err() != nil {
				return results, ctx.Err()
			}
			if err != nil {
				result.Content = "tool execution failed: " + err.Error()
			} else {
				result = output
				result.ToolCallID = call.ID
				if result.Status != tools.StatusCompleted {
					return results, fmt.Errorf("unexpected tool result status %q for call %s", result.Status, call.ID)
				}
			}
		default:
			return results, fmt.Errorf("unknown safety decision %q for call %s", decision.Action, call.ID)
		}
		if err := e.recorder.Record(ctx, agentloop.Entry{
			Type: agentloop.EntryTypeToolResult,
			Data: map[string]any{
				"tool_call_id": result.ToolCallID,
				"tool":         call.Name,
				"status":       result.Status,
				"content":      result.Content,
			},
		}); err != nil {
			return results, fmt.Errorf("record tool result for call %s: %w", call.ID, err)
		}
		results = append(results, result)
		// Loop 负责追加已完成结果，遇到等待项时通过 Recorder 记录暂停，否则继续调用模型。

	}
	return results, nil
}
