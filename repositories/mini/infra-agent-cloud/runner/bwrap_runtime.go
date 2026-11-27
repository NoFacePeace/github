package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"infra-agent-cloud/executor"
)

// BwrapRuntime 在 Bubblewrap 沙箱中运行 Executor。
type BwrapRuntime struct {
	DataRoot           string
	BwrapExecutable    string
	ExecutorExecutable string
	// RootlessKitExecutable 为空时沿用宿主机网络。
	RootlessKitExecutable string
	CgroupRoot            string
	Limits                BwrapResourceLimits
	// StopGracePeriod 为零时使用 5 秒停止宽限期。
	StopGracePeriod time.Duration
}

const (
	bwrapTaskRoot     = "/task"
	bwrapExecutorPath = "/opt/executor"
)

// bwrapExecutionPaths 保存宿主机上的目录和执行清单路径。
type bwrapExecutionPaths struct {
	TaskRoot  string
	Workspace string
	Config    string
	Session   string
	Tmp       string
	Manifest  string
}

// Start 准备沙箱并启动本轮 Executor：
//  1. 准备工作区和本轮临时目录，写入不含凭证的执行清单。
//  2. 准备沙箱挂载、环境变量和 seccomp 配置，组装沙箱内的 Executor 启动命令。
//  3. 创建本轮 cgroup 并设置资源限制，按需配置 RootlessKit 网络隔离。
//  4. 启动包装进程，先加入 cgroup，再由 Bubblewrap 在沙箱内启动 Executor。
//  5. 通过 stdin 传递凭证。普通模式写完后关闭 stdin，steer 模式保留输入管道。
//  6. 返回执行句柄供调用方读取输出和等待退出。启动失败时清理由本轮创建的资源。
func (r *BwrapRuntime) Start(ctx context.Context, spec *executor.Spec, secrets *executor.Secrets) (handle *ExecutionHandle, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	paths, err := r.prepareExecutionPaths(spec)
	if err != nil {
		return nil, err
	}
	artifacts := bwrapArtifacts{manifest: paths.Manifest}
	defer func() {
		if err != nil {
			err = errors.Join(err, artifacts.remove())
		}
	}()
	seccompFile, err := r.writeSeccompFile()
	if err != nil {
		return nil, err
	}
	artifacts.seccomp = seccompFile
	artifacts.childPID = seccompFile + ".pid"
	cgroup, err := r.createExecutionCgroup(spec.ExecutionID)
	if err != nil {
		return nil, err
	}
	artifacts.cgroup = cgroup
	networkDir, err := r.prepareExecutionNetwork()
	if err != nil {
		return nil, err
	}
	artifacts.network = networkDir
	// Start 的 Context 控制启动，启动成功后的停止由 Runtime 管理。
	cmd, err := r.buildBwrapCommand(context.WithoutCancel(ctx), paths, seccompFile, cgroup, networkDir, artifacts.childPID)
	if err != nil {
		return nil, err
	}
	handle, err = startBwrapProcess(ctx, cmd, spec, secrets, cgroup, artifacts.childPID)
	if err != nil {
		return nil, err
	}
	handle.bwrapArtifacts = artifacts
	return handle, nil
}

type bwrapArtifacts struct {
	manifest string
	seccomp  string
	cgroup   string
	network  string
	childPID string
}

func (a bwrapArtifacts) remove() error {
	var cleanupErr error
	if a.network != "" {
		cleanupErr = os.RemoveAll(a.network)
	}
	for _, name := range []string{a.cgroup, a.childPID, a.seccomp, a.manifest} {
		if name != "" {
			if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
				cleanupErr = errors.Join(cleanupErr, err)
			}
		}
	}
	return cleanupErr
}

// buildBwrapCommand 组装沙箱挂载和 Executor 启动参数。
// 启用 RootlessKit 时，完整启动命令如下，尖括号内为运行时参数：
//
//	/bin/sh -c 'set -eu; printf "%s\n" "$$" > "$1/cgroup.procs"; shift; exec "$@"' \
//	  bwrap-cgroup <cgroup> \
//	  <rootlesskit> --net=slirp4netns --disable-host-loopback --copy-up=/etc \
//	    --port-driver=none --state-dir=<networkDir>/rootlesskit \
//	  /bin/sh -c 'set -eu; umask 077; exec 3>"$1" 4<"$2"; shift 2; exec "$@"' \
//	  bwrap-setup <childPIDFile> <seccompFile> \
//	  <bwrap> --die-with-parent --new-session \
//	    --unshare-user --uid 0 --gid 0 --unshare-pid --unshare-uts \
//	    --cap-drop ALL --info-fd 3 --seccomp 4 \
//	    --proc /proc --dev /dev --tmpfs /tmp --dir /opt \
//	    <系统路径只读挂载> \
//	    --bind <taskRoot> /task --ro-bind <executorPath> /opt/executor \
//	    --clearenv --setenv PATH /usr/local/bin:/usr/bin:/bin \
//	    --setenv HOME /tmp --setenv TMPDIR /tmp --chdir /task/workspace \
//	    -- /opt/executor -command=run -manifest=/task/tmp/<ExecutionID>.json
//
// 未启用 RootlessKit 时，cgroup shell 直接启动 setup shell。
// 系统路径只读挂载按宿主机上存在的路径展开为 --ro-bind 参数。
// 凭证 JSON 通过 stdin 传给 Executor。
func (r *BwrapRuntime) buildBwrapCommand(ctx context.Context, paths bwrapExecutionPaths, seccompFile, cgroup, networkDir, childPIDFile string) (*exec.Cmd, error) {
	bwrap, err := exec.LookPath(r.BwrapExecutable)
	if err != nil {
		return nil, fmt.Errorf("resolve bwrap executable: %w", err)
	}
	executorPath, err := exec.LookPath(r.ExecutorExecutable)
	if err != nil {
		return nil, fmt.Errorf("resolve executor executable: %w", err)
	}
	executorPath, err = filepath.Abs(executorPath)
	if err != nil {
		return nil, fmt.Errorf("resolve executor path: %w", err)
	}

	args := []string{
		"--die-with-parent", "--new-session",
		"--unshare-user", "--uid", "0", "--gid", "0", "--unshare-pid", "--unshare-uts",
		"--cap-drop", "ALL", "--info-fd", "3", "--seccomp", "4",
		"--proc", "/proc", "--dev", "/dev", "--tmpfs", "/tmp", "--dir", "/opt",
	}
	for _, systemPath := range []string{
		"/usr", "/bin", "/lib", "/lib64", "/etc/alternatives",
		"/etc/nsswitch.conf", "/etc/resolv.conf", "/etc/ssl", "/etc/ld.so.cache",
	} {
		if _, err := os.Stat(systemPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("check system mount %q: %w", systemPath, err)
		}
		args = append(args, "--ro-bind", systemPath, systemPath)
	}
	args = append(args,
		"--bind", paths.TaskRoot, bwrapTaskRoot,
		"--ro-bind", executorPath, bwrapExecutorPath,
		"--clearenv", "--setenv", "PATH", "/usr/local/bin:/usr/bin:/bin",
		"--setenv", "HOME", "/tmp", "--setenv", "TMPDIR", "/tmp",
		"--chdir", path.Join(bwrapTaskRoot, "workspace"),
		"--", bwrapExecutorPath, "-command=run", "-manifest="+path.Join(bwrapTaskRoot, "tmp", filepath.Base(paths.Manifest)),
	)
	// 在网络包装内打开信息和 seccomp 文件，避免依赖外层保留额外的文件描述符。
	launchArgs := []string{"-c", `set -eu; umask 077; exec 3>"$1" 4<"$2"; shift 2; exec "$@"`, "bwrap-setup", childPIDFile, seccompFile, bwrap}
	command := "/bin/sh"
	launchArgs = append(launchArgs, args...)
	if networkDir != "" {
		rootlessKit, err := exec.LookPath(r.RootlessKitExecutable)
		if err != nil {
			return nil, fmt.Errorf("resolve rootlesskit executable: %w", err)
		}
		networkArgs := []string{
			"--net=slirp4netns", "--disable-host-loopback", "--copy-up=/etc", "--port-driver=none",
			"--state-dir=" + filepath.Join(networkDir, "rootlesskit"), command,
		}
		command, launchArgs = rootlessKit, append(networkArgs, launchArgs...)
	}
	return inExecutionCgroup(ctx, cgroup, command, launchArgs), nil
}

func (r *BwrapRuntime) prepareExecutionPaths(spec *executor.Spec) (bwrapExecutionPaths, error) {
	if r.DataRoot == "" {
		return bwrapExecutionPaths{}, errors.New("bwrap data root is required")
	}
	for _, id := range []string{spec.TaskID, spec.ExecutionID} {
		if !filepath.IsLocal(id) || id == "." || filepath.Base(id) != id {
			return bwrapExecutionPaths{}, errors.New("invalid bwrap task or execution ID")
		}
	}
	dataRoot, err := filepath.Abs(r.DataRoot)
	if err != nil {
		return bwrapExecutionPaths{}, fmt.Errorf("resolve bwrap data root: %w", err)
	}
	taskRoot := filepath.Join(dataRoot, "tasks", spec.TaskID)
	paths := bwrapExecutionPaths{
		TaskRoot:  taskRoot,
		Workspace: filepath.Join(taskRoot, "workspace"),
		Config:    filepath.Join(taskRoot, "config"),
		Session:   filepath.Join(taskRoot, "session"),
		Tmp:       filepath.Join(taskRoot, "tmp"),
	}
	for _, dir := range []string{paths.Workspace, paths.Config, paths.Session, paths.Tmp} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return bwrapExecutionPaths{}, fmt.Errorf("prepare bwrap directory: %w", err)
		}
	}

	manifest := *spec
	manifest.Paths = executor.Paths{
		Workspace: path.Join(bwrapTaskRoot, "workspace"),
		Config:    path.Join(bwrapTaskRoot, "config"),
		Session:   path.Join(bwrapTaskRoot, "session"),
	}
	paths.Manifest = filepath.Join(paths.Tmp, spec.ExecutionID+".json")
	file, err := os.OpenFile(paths.Manifest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return bwrapExecutionPaths{}, fmt.Errorf("create execution manifest: %w", err)
	}
	writeErr := json.NewEncoder(file).Encode(manifest)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return bwrapExecutionPaths{}, fmt.Errorf("write execution manifest: %w", errors.Join(err, os.Remove(paths.Manifest)))
	}
	return paths, nil
}
