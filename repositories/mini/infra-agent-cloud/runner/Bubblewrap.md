# Bubblewrap 启动示例

## 1. 从 cgroup 到 Executor 的完整启动命令

这条 Bash 命令串联 cgroup、RootlessKit、防火墙、bwrap 和 Node.js Executor。运行前，启动管理程序应已创建本轮 cgroup 并写入 `memory.max`、`cpu.max`、`pids.max`，准备网络状态目录、任务目录、执行清单及 Git 配置。按实际环境替换示例路径，并将 `host_ipv4` 换成宿主机的 IPv4 地址列表。需要加载 seccomp 时，可在内层 shell 打开 FD 4，并向 bwrap 传入 `--seccomp 4`。

```bash
task_root=/var/lib/sandbox-runtime/tasks/task-demo                 # 宿主机上的任务目录。
network_dir=/var/lib/sandbox-runtime/networks/execution-demo      # 本轮网络状态和 PID 文件目录。
cgroup=/sys/fs/cgroup/sandbox-runtime/execution-demo              # 已经写入资源限制的执行 cgroup。
node_executable=/usr/bin/node                                    # 宿主机上的 Node.js 可执行文件。
executor_file=/opt/sandbox-runtime/executor.js                    # 宿主机上的 Executor 脚本。
execution_id=execution-demo                                       # 对应本轮 manifest 的文件名。

host_ipv4=(127.0.0.1 192.0.2.10) # 宿主机 IPv4 地址示例，可用 ip -j -4 address show 收集并去重。

system_paths=( # 列出候选系统路径。
  /usr /bin /lib /lib64 /etc/alternatives                         # 系统程序、动态库及命令链接。
  /etc/hosts /etc/nsswitch.conf /etc/resolv.conf                   # 主机名和 DNS 解析配置。
  /etc/ssl /etc/ld.so.cache                                       # TLS 证书和动态链接器缓存。
) # 结束系统路径列表。
system_mounts=() # 保存实际传给 bwrap 的只读挂载参数。
for system_path in "${system_paths[@]}"; do                          # 逐项处理候选路径。
  if [ -e "$system_path" ]; then                                     # 跳过宿主机上不存在的路径。
    system_mounts+=(--ro-bind "$system_path" "$system_path")          # 以相同路径只读挂入沙箱。
  fi                                                                 # 结束路径存在检查。
done                                                                 # 结束挂载参数组装。

cgroup_script=' # 外层 shell 在启动隔离环境前执行这段脚本。
  set -eu                                       # 命令失败或变量未设置时退出。
  printf "%s\n" "$$" > "$1/cgroup.procs"         # 将当前 shell 的 PID 加入本轮 cgroup。
  shift                                         # 移除已消费的 cgroup 目录参数。
  exec "$@"                                     # 以原 PID 启动 RootlessKit，并传递剩余参数。
' # 结束外层 shell 脚本。

network_script=' # 在 RootlessKit 创建的网络命名空间内执行这段脚本。
  set -eu                                      # 网络规则设置失败或变量未设置时退出。
  pid_file=$1                                  # 记录 bwrap 子进程信息的文件路径。
  shift                                        # 移除 PID 文件参数，接下来是 IPv4 拒绝列表。
  ip6tables -w -A OUTPUT ! -o lo -j REJECT       # 拒绝经非 loopback 接口发出的 IPv6 流量。
  while [ "$1" != -- ]; do                      # 逐个读取目的地址，直到遇到参数分隔符。
    iptables -w -A OUTPUT ! -o lo -d "$1" -j REJECT # 拒绝经非 loopback 接口发往该 IPv4 地址的流量。
    shift                                       # 处理下一个地址。
  done                                          # 结束 IPv4 规则配置。
  shift                                         # 移除分隔符，剩余参数是 bwrap 命令。
  exec 3>"$pid_file"                             # 打开 FD 3，供 bwrap 写出子进程信息。
  exec "$@"                                      # 启动 bwrap，并让它继承已配置的网络和 FD 3。
' # 结束网络脚本。

cd "$task_root/workspace" || exit 1 # 在宿主机侧进入任务工作区，失败时退出。

launch=( # 逐项组装完整命令并保留每个参数的边界。
  env -i                                                           # 清空启动器继承的环境。
  PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin  # 为外层工具设置搜索路径。
  XTABLES_LOCKFILE="$network_dir/xtables.lock"                       # 为本轮防火墙命令指定锁文件。
  /bin/sh -c "$cgroup_script" sandbox-cgroup "$cgroup"               # 执行 cgroup 脚本，标签是 $0，目录是 $1。
  rootlesskit                                                      # 启动外层隔离环境。
  --net=slirp4netns                                                 # 使用 slirp4netns 提供网络连接。
  --disable-host-loopback                                           # 禁止通过该网络访问宿主机 loopback。
  --copy-up=/etc                                                    # 为隔离环境准备私有的 /etc 挂载。
  --port-driver=none                                                # 关闭 RootlessKit 的入站端口转发驱动。
  --state-dir="$network_dir/rootlesskit"                             # 保存本轮 RootlessKit 状态。
  /bin/sh -c "$network_script" sandbox-network "$network_dir/child-pid.json" # 启动网络脚本，PID 文件路径是 $1。
  0.0.0.0/8 "${host_ipv4[@]}" --                                     # 传入 IPv4 拒绝列表，并以 -- 标记列表结束。
  bwrap                                                            # 在规则设置完成后启动 Bubblewrap。
  --die-with-parent                                                # 父进程退出时结束 Bubblewrap 沙箱。
  --new-session                                                    # 为沙箱创建独立的进程会话。
  --unshare-user --uid 0 --gid 0                                    # 创建用户命名空间，内部身份使用 UID 和 GID 0。
  --unshare-pid                                                    # 创建独立的 PID 命名空间。
  --proc /proc --dev /dev                                           # 挂载 proc 文件系统并建立基本设备目录。
  --dir /opt --dir /sandbox --dir /tmp                              # 准备二进制、任务及临时目录的挂载位置。
  --ro-bind "$node_executable" /opt/node                             # 只读挂载 Node.js。
  --ro-bind "$executor_file" /opt/executor.js                       # 只读挂载 Executor 脚本。
  --info-fd 3                                                       # 通过 FD 3 输出 Bubblewrap 子进程信息。
  --cap-drop ALL                                                    # 丢弃沙箱内的全部 Linux capabilities。
  "${system_mounts[@]}"                                             # 展开实际存在的系统只读挂载。
  --bind "$task_root" /sandbox                                     # 将任务目录可写挂入沙箱。
  --bind "$task_root/tmp" /tmp                                     # 让沙箱内的 /tmp 指向任务临时目录。
  --clearenv                                                        # 清空传给沙箱内目标程序的环境。
  --setenv PATH /sandbox/home/.sandbox-git:/usr/local/bin:/usr/bin:/bin # 让 Executor 优先找到准备好的 Git 包装程序。
  --setenv HOME /sandbox/home                                       # 设置沙箱内的用户目录。
  --setenv TMPDIR /sandbox/tmp                                     # 设置任务的临时文件目录。
  --setenv XDG_CACHE_HOME /sandbox/home/.cache                        # 指定用户缓存目录。
  --setenv XDG_CONFIG_HOME /sandbox/home/.config                      # 指定用户配置目录。
  --setenv XDG_DATA_HOME /sandbox/home/.local/share                   # 指定用户数据目录。
  --setenv GIT_CONFIG_GLOBAL /sandbox/home/.sandbox-git/gitconfig    # 使用为任务准备的 Git 配置。
  --setenv GIT_CONFIG_NOSYSTEM 1                                    # 跳过系统级 Git 配置。
  --setenv GIT_TERMINAL_PROMPT 0                                    # 禁止 Git 通过终端交互询问凭证。
  --chdir /sandbox/workspace                                        # 在沙箱内进入任务工作区。
  /opt/node /opt/executor.js run --manifest "/sandbox/tmp/$execution_id.json" # 让 Node.js 启动 Executor 并读取本轮清单。
) # 结束命令数组。

"${launch[@]}" # 执行完整启动命令，保留调用方连接的 stdin、stdout 和 stderr。
```

外层 shell 先把自己的 PID 写入 `cgroup.procs`，随后启动的进程继承该 cgroup。网络规则在 RootlessKit 的网络命名空间内安装，FD 3 由 bwrap 用来写出子进程信息。凭证 JSON 由启动管理程序通过 stdin 传入。普通模式写完后关闭 stdin，steer 模式保留该管道。
