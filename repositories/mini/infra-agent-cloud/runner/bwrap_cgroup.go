package runner

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// BwrapResourceLimits 定义每次执行的 cgroup v2 资源上限。
type BwrapResourceLimits struct {
	MemoryBytes int64
	// CPUQuotaMicros 是每 100000 微秒可用的 CPU 时间，100000 表示一个核心。
	CPUQuotaMicros int64
	PIDs           int64
}

func (r *BwrapRuntime) createExecutionCgroup(executionID string) (string, error) {
	if !filepath.IsAbs(r.CgroupRoot) {
		return "", errors.New("bwrap cgroup root must be an absolute path")
	}
	if r.Limits.MemoryBytes <= 0 || r.Limits.CPUQuotaMicros < 1000 || r.Limits.PIDs <= 0 {
		return "", errors.New("invalid bwrap resource limits")
	}
	if _, err := os.Stat(filepath.Join(r.CgroupRoot, "cgroup.controllers")); err != nil {
		return "", fmt.Errorf("check bwrap cgroup v2 root: %w", err)
	}
	dir := filepath.Join(r.CgroupRoot, fmt.Sprintf("execution-%x", sha256.Sum256([]byte(executionID))))
	if err := os.Mkdir(dir, 0o700); err != nil {
		return "", fmt.Errorf("create execution cgroup: %w", err)
	}
	if err := writeBwrapResourceLimits(dir, r.Limits); err != nil {
		return "", errors.Join(err, os.Remove(dir))
	}
	return dir, nil
}

func writeBwrapResourceLimits(dir string, limits BwrapResourceLimits) error {
	for _, control := range []struct{ name, value string }{
		{"memory.max", strconv.FormatInt(limits.MemoryBytes, 10)},
		{"cpu.max", strconv.FormatInt(limits.CPUQuotaMicros, 10) + " 100000"},
		{"pids.max", strconv.FormatInt(limits.PIDs, 10)},
	} {
		// 只打开内核已提供的控制文件，控制器不可用时中止启动。
		file, err := os.OpenFile(filepath.Join(dir, control.name), os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("open cgroup %s: %w", control.name, err)
		}
		_, writeErr := io.WriteString(file, control.value)
		if err := errors.Join(writeErr, file.Close()); err != nil {
			return fmt.Errorf("write cgroup %s: %w", control.name, err)
		}
	}
	return nil
}

// inExecutionCgroup 在执行沙箱命令前将包装进程加入本轮 cgroup。
func inExecutionCgroup(ctx context.Context, dir, command string, args []string) *exec.Cmd {
	launchArgs := []string{"-c", `set -eu; printf '%s\n' "$$" > "$1/cgroup.procs"; shift; exec "$@"`, "bwrap-cgroup", dir, command}
	return exec.CommandContext(ctx, "/bin/sh", append(launchArgs, args...)...)
}
