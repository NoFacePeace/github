# 准备启动文件与资源限制

读取执行资源限制，准备任务目录与 manifest，生成 bwrap 需要的 hosts 和 seccomp 辅助文件。

## 1. 文件准备

任务目录包括 workspace、config、session、memory、home 和 tmp。manifest 保存执行描述，不放入完整敏感输入。

hosts 与 seccomp 等私有辅助文件放在任务挂载范围外的进程管理目录，并设置受限权限；普通执行不能直接访问这些宿主管理文件。

## 2. 资源限制

准备适用的任务 cgroup 根目录，再为本次执行创建 cgroup，应用执行资源限制。cgroup 管理进程资源，与 Redis 中预留 Runner 并发名额的容量集合不是同一机制。

## 3. systemd 与管理权限

VM 部署由 systemd 启动 Runner。systemd 是 Linux 服务管理程序，负责服务启停、重启和运行环境；cgroup 则是内核提供的资源控制机制。

服务的 `[Service]` 配置预先委派 CPU、内存和进程数控制器，并将 Runner 放入单独的 supervisor 子分组。以下分组名为通用示例：

```ini
[Service]
Delegate=cpu memory pids
DelegateSubgroup=runner-supervisor
```

`Delegate` 允许服务管理其下的资源控制分组；`DelegateSubgroup` 安排服务自身所在的子分组。二者本身不设置具体 CPU、内存或进程数上限，仍需结合服务及父级 cgroup 的资源约束。

Runner 使用已有委派权限，不是在启动任务时临时请求 systemd 授权。部署文件存在不代表目标机器已加载生效。

## 4. 确定 cgroup 根目录

优先使用显式配置的 cgroup 根目录。未配置时，读取 `/proc/self/cgroup` 中 `0::` 开头的行，取得当前进程所属 cgroup，再取 supervisor 所在路径的父目录，拼接到 `/sys/fs/cgroup` 下。

以通用服务名称举例：

```text
/proc/self/cgroup：
0::/system.slice/runner.service/runner-supervisor

推导的管理根目录：
/sys/fs/cgroup/system.slice/runner.service/
```

实际实现会校验受支持的 supervisor 分组名称。随后检查根目录为绝对路径、文件系统为 cgroup v2，解析真实路径，并向 `cgroup.subtree_control` 写入 `+cpu +memory +pids`，启用对子分组的控制能力。

默认 VM 部署依赖 systemd 委派后自动推导；配置覆盖时优先使用指定目录。没有满足条件的目录或写权限时失败，不默默跳过资源限制。

## 5. 分组结构

未启用任务池准入时，每轮执行分组与 Runner 自身分组同级。以下名称均为示意：

```text
/sys/fs/cgroup/system.slice/runner.service/
├── runner-supervisor/          # Runner 自身
├── execution-<执行哈希A>/
└── execution-<执行哈希B>/
```

启用任务池准入后，在管理根目录下增加任务池，再在池内创建各轮执行分组：

```text
服务 cgroup 根目录/
├── runner-supervisor/
└── tasks/                     # 任务池整体内存预算
    ├── execution-<执行哈希A>/
    └── execution-<执行哈希B>/
```

代码缺省关闭任务池准入；部署是否启用取决于运行配置。执行组不建在 Runner 的 supervisor 分组内部。子组仍受父组资源上限约束。

## 6. 创建执行分组

根据本轮 `executionId` 计算 SHA256，生成带固定前缀的执行组名称，再用文件系统 API 创建目录，不复用已有执行组。内核会在新 cgroup 中提供控制文件。

随后写入限制，代码缺省值为：

| 控制文件 | 缺省写入值 | 含义 |
| --- | --- | --- |
| memory.max | `4294967296` | 内存上限 4 GiB。 |
| memory.swap.max | `0` | 不允许使用 swap。 |
| memory.oom.group | `1` | 对本执行组启用整体 OOM 处理语义。 |
| cpu.max | `400000 100000` | 每 100ms 周期最多使用 400ms CPU 时间，约 4 核配额。 |
| pids.max | `512` | 限制任务数量，包括线程。 |

这些值可由执行资源配置调整。写入控制文件是配置内核约束，不是普通磁盘配置保存。最后检查 `cgroup.kill` 是否可写；设置失败则移除本次新建分组并报错。

创建分组时 Executor 尚未运行。后续启动脚本再将进程加入 `cgroup.procs`，此后才对组内进程施加约束。

## 7. 后续步骤

进入[准备 MCP 与 Git 环境](PrepareToolEnvironment.md)。
