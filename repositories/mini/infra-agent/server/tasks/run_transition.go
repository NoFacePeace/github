package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// updateRunStatus 组装状态事件，交给存储层在同一事务中校验租约并转换状态。
func (s *service) updateRunStatus(ctx context.Context, task Task, item run, credential *lease, status RunStatus, reason string) (bool, error) {
	from := []RunStatus{RunStatusPending, RunStatusRunning, RunStatusWaitingForApproval, RunStatusWaitingForInput}
	var eventType EventType
	switch status {
	case RunStatusRunning:
		from = []RunStatus{item.Status}
		if item.Status == RunStatusPending {
			eventType = EventRunStarted
		} else {
			eventType = EventRunResumed
		}
	case RunStatusFailed:
		eventType = EventRunFailed
	case RunStatusCancelled:
		eventType = EventRunCancelled
	default:
		return false, fmt.Errorf("unsupported run status transition to %q", status)
	}
	eventID, err := newID("event")
	if err != nil {
		return false, fmt.Errorf("generate status event id: %w", err)
	}
	payload, err := json.Marshal(map[string]string{"status": string(status), "reason": reason})
	if err != nil {
		return false, fmt.Errorf("encode status event: %w", err)
	}
	changed, err := s.store.UpdateRunStatus(ctx, item.ID, UpdateRunStatusRequest{
		From: from, To: status, Owner: credential.Owner, Epoch: credential.Epoch,
		Event: Event{
			ID: eventID, TaskID: task.ID, RunID: item.ID, Type: eventType,
			Actor: "runtime", Payload: payload, CreatedAt: time.Now().UTC(),
		},
	})
	if err != nil {
		return false, fmt.Errorf("update run %q status to %q: %w", item.ID, status, err)
	}
	return changed, nil
}
