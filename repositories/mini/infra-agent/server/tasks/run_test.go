package tasks

import "testing"

func TestRunStatusIsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status RunStatus
		want   bool
	}{
		{name: "pending", status: RunStatusPending},
		{name: "running", status: RunStatusRunning},
		{name: "waiting for input", status: RunStatusWaitingForInput},
		{name: "waiting for approval", status: RunStatusWaitingForApproval},
		{name: "completed", status: RunStatusCompleted, want: true},
		{name: "failed", status: RunStatusFailed, want: true},
		{name: "cancelled", status: RunStatusCancelled, want: true},
		{name: "empty"},
		{name: "unknown", status: RunStatus("unknown")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}
