package bwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"infra-agent-cloud/runner"
)

func readBwrapChildPID(ctx context.Context, filename string, exited <-chan struct{}) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	finished := false
	for {
		data, err := os.ReadFile(filename)
		if err == nil {
			var info struct {
				PID int `json:"child-pid"`
			}
			if json.Unmarshal(data, &info) == nil {
				if info.PID <= 1 {
					return 0, errors.New("bwrap reported an invalid child PID")
				}
				return info.PID, nil
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("read bwrap child PID: %w", err)
		}
		if finished {
			return 0, errors.New("bwrap exited before reporting its child PID")
		}
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("wait for bwrap child PID: %w", context.Cause(ctx))
		case <-exited:
			// 退出通知和最后一次文件写入可能同时到达，再检查一次。
			finished = true
		case <-ticker.C:
		}
	}
}

// Stop 请求沙箱内的执行进程停止，宽限期后通过 cgroup 终止剩余进程。
func (r *BwrapRuntime) Stop(ctx context.Context, handle *runner.ExecutionHandle) error {
	if err := context.Cause(ctx); err != nil {
		return err
	}
	state, ok := r.executions.Load(handle)
	if !ok || handle.WaitForExit == nil {
		return errors.New("invalid bwrap execution handle")
	}
	execution := state.(bwrapExecution)
	select {
	case <-execution.exited:
		_, err := handle.WaitForExit(ctx)
		return err
	default:
	}

	interruptErr := interruptBwrapProcess(handle, execution.artifacts.cgroup)
	if interruptErr == nil {
		grace := r.StopGracePeriod
		if grace <= 0 {
			grace = 5 * time.Second
		}
		graceCtx, cancel := context.WithTimeout(ctx, grace)
		_, err := handle.WaitForExit(graceCtx)
		timedOut := errors.Is(graceCtx.Err(), context.DeadlineExceeded)
		cancel()
		if err == nil || !timedOut {
			return err
		}
	}
	if err := context.Cause(ctx); err != nil {
		return errors.Join(interruptErr, err)
	}
	if err := killBwrapCgroup(execution.artifacts.cgroup); err != nil {
		return errors.Join(interruptErr, fmt.Errorf("kill execution cgroup: %w", err))
	}
	_, err := handle.WaitForExit(ctx)
	return err
}

func interruptBwrapProcess(handle *runner.ExecutionHandle, dir string) error {
	pid := handle.InterruptPID
	if pid <= 1 || dir == "" {
		return errors.New("missing bwrap process identity")
	}
	// 仅向仍在本轮 cgroup 中的 PID 发信号，避免操作已复用的 PID。
	members, err := os.ReadFile(filepath.Join(dir, "cgroup.procs"))
	if err != nil {
		return fmt.Errorf("read execution cgroup membership: %w", err)
	}
	if !slices.Contains(strings.Fields(string(members)), strconv.Itoa(pid)) {
		return fmt.Errorf("bwrap process %d is not in the execution cgroup", pid)
	}
	if !handle.InterruptProcessOnly {
		if err := syscall.Kill(-pid, syscall.SIGINT); !errors.Is(err, syscall.ESRCH) {
			return err
		}
	}
	return syscall.Kill(pid, syscall.SIGINT)
}
