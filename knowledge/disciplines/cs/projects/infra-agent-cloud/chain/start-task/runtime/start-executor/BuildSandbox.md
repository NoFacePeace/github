# 构建 bwrap 隔离

根据任务路径、工具挂载与环境信息生成 bwrap 启动参数，并加载已生成的 seccomp 策略。

## 1. 文件与进程环境

限定任务可访问的文件目录，将 Executor 脚本与 Node.js 等必要资源挂入沙箱，配置工作目录、环境变量及受控的系统文件。任务通过沙箱内路径读取 manifest。

## 2. 网络隔离

按当前远程执行路径的隔离网络配置，使用 RootlessKit 与 slirp4netns 等准备独立网络环境及出站访问规则。网络配置准备成功后，取得实际需要启动的命令与参数。

工具分工、启动选项以及子命令的执行关系见 [RootlessKit 网络环境](RootlessKit.md)。

沙箱、网络及 seccomp 限制属于互补机制，不能只因使用 bwrap 就假设所有访问均被阻断。

## 3. 系统调用过滤

bwrap 通过 `--seccomp 4` 读取并安装启动管理程序生成的 BPF 过滤器。一份精简策略可默认允许其余调用，明确拒绝 14 个系统调用，并校验原生系统调用 ABI；示例清单、选择依据和加载过程见[Seccomp 执行策略](SeccompPolicy.md)。

## 4. 启动命令模板

下面按“外层 RootlessKit 网络 + 内层 bwrap + Node.js Executor”的组合展开参数。路径和主机名使用通用名称，这是调用链的配置示例，不是 bwrap 的标准默认启动命令。

### 4.1 启动前提

启动管理程序已经准备好任务目录、manifest、Node.js、Executor 文件、hosts 文件和 seccomp BPF 文件，并根据工具需求生成挂载及环境变量参数。

外层启动脚本先将执行进程加入 cgroup，RootlessKit 建立网络环境，网络脚本配置规则，再由包装脚本打开描述符：

```sh
exec 3>"$pid_file"      # 网络脚本：准备子进程信息输出通道
exec 4<"$seccomp_file"  # seccomp 包装脚本：准备过滤器输入通道
```

以上展示描述符准备顺序；实际调用链中，两行可以位于先后通过 `exec` 执行的不同脚本中。

### 4.2 完整参数结构

```sh
exec bwrap \
  --die-with-parent \
  --new-session \
  --unshare-user \
  --uid 0 \
  --gid 0 \
  --unshare-pid \
  --unshare-uts \
  --hostname sandbox \
  --proc /proc \
  --dev /dev \
  --dir /opt \
  --dir /sandbox \
  --dir /tmp \
  --ro-bind "$node_executable" /opt/node \
  --ro-bind "$executor_file" /opt/executor.js \
  --info-fd 3 \
  --cap-drop ALL \
  --seccomp 4 \
  "${system_mount_args[@]}" \
  --ro-bind "$hosts_file" /etc/hosts \
  --bind "$task_root" /sandbox \
  --bind "$task_tmp" /tmp \
  --clearenv \
  "${mcp_mount_args[@]}" \
  "${env_args[@]}" \
  --chdir /sandbox/workspace \
  /opt/node /opt/executor.js run \
  --manifest "/sandbox/tmp/$manifest_name"
```

这是需要填充变量的 Bash 数组模板，不是可直接复制执行的独立命令。外层包装脚本也可以使用 POSIX Shell，但应由启动管理程序按参数列表传入已展开的命令，不能直接在 POSIX Shell 中使用这里的数组语法。

示例假设 `$task_root` 下包含 `workspace`、`tmp` 等目录，`$task_tmp` 对应其中的任务临时目录，manifest 已写入该临时目录。挂载后，沙箱内才能通过 `/sandbox/tmp/$manifest_name` 读取它。

### 4.3 动态参数

| 参数或变量 | 展开方式与含义 |
| --- | --- |
| `$node_executable` | 宿主上的 Node.js 可执行文件，映射到 `/opt/node`。 |
| `$executor_file` | 宿主上的 Executor 脚本，映射到 `/opt/executor.js`。 |
| `system_mount_args` | 为存在且需要的系统路径追加 `--ro-bind 源路径 目标路径`。 |
| `$hosts_file` | 提前生成的专用 hosts 文件，只读映射到 `/etc/hosts`。 |
| `$task_root`、`$task_tmp` | 当前任务可写目录和临时目录。 |
| `mcp_mount_args` | 为需要的 MCP 工具文件或目录追加只读挂载参数。 |
| `env_args` | 为每个受控变量追加 `--setenv 名称 值`。 |
| `$manifest_name` | 本次执行描述文件的文件名。 |

系统只读挂载的候选路径包括 `/usr`、`/bin`、`/lib`、`/lib64`、`/etc/alternatives`、`/etc/nsswitch.conf`、`/etc/resolv.conf`、`/etc/ssl`、`/etc/ld.so.cache`。启动管理程序按宿主路径是否存在及任务实际需要选择；这些路径不是 bwrap 自动添加的默认清单。

环境变量需要使用沙箱内路径。例如，任务 HOME 可设为 `/sandbox/home`，临时目录可设为 `/tmp`。`--clearenv` 只清理环境变量，不负责清除已打开的文件描述符或其他进程状态。

## 5. 参数的作用与自定义约定

| 选项 | 作用 |
| --- | --- |
| `--die-with-parent` | 请求在父进程退出时终止相关沙箱进程，配合外层进程及 cgroup 回收使用。 |
| `--new-session` | 创建新会话，隔离终端会话关系。 |
| `--unshare-user --uid 0 --gid 0` | 创建内层用户命名空间，内部使用 UID/GID 0；不授予宿主 root 权限。 |
| `--unshare-pid` | 创建独立 PID 命名空间。 |
| `--unshare-uts --hostname sandbox` | 创建独立 UTS 命名空间并设置主机名。 |
| `--proc /proc`、`--dev /dev` | 准备 proc 文件系统和基础设备环境。 |
| `--dir` | 在沙箱文件系统中创建指定目录。 |
| `--ro-bind`、`--bind` | 分别配置只读和可写绑定挂载；可写绑定中的修改会作用于对应宿主目录。 |
| `--info-fd 3` | 向已打开的 FD 3 输出子进程信息。 |
| `--cap-drop ALL` | 移除全部 capabilities。 |
| `--seccomp 4` | 从已打开的 FD 4 读取并安装过滤器。 |
| `--clearenv`、`--setenv` | 清理继承环境，再设置指定变量。 |
| `--chdir` | 设置目标程序工作目录。 |

这些选项属于 bwrap 的通用接口；是否使用、挂载哪些目录以及启用哪些 namespace，由调用方决定。FD 3/4 的编号、`/sandbox` 目录布局、主机名、Node.js Executor 和 manifest 参数均是本示例的约定，可根据目标程序调整。

## 6. 网络继承与目标命令

模板没有传入 `--unshare-net`，因此 bwrap 沿用外层 RootlessKit 已创建并配置的网络命名空间。内层再创建 user namespace 并移除 capabilities，使任务使用既有网络，同时不具备修改外层网络规则所需的权限。

bwrap 的通用语法为：

```text
bwrap [隔离与挂载选项] 目标命令 [目标命令参数...]
```

模板从 `/opt/node` 开始是目标命令，`run --manifest ...` 是 Executor 的参数，不是 bwrap 的选项：

```text
外层包装脚本
  └─ exec bwrap：读取参数，构建沙箱并安装过滤器
       └─ /opt/node /opt/executor.js run --manifest ...
            └─ Executor 启动任务使用的 CLI、MCP 及其他子进程
```

组装参数发生在宿主侧准备阶段；真正创建沙箱和执行目标命令发生在后续进程启动阶段。完整隔离还依赖外层网络规则、cgroup、凭证传递和收尾机制，不能只凭这条 bwrap 命令判断所有边界均已生效。

## 7. 后续步骤

进入[创建 Executor 进程](SpawnExecutor.md)。
