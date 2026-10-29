# Namespace

Linux namespace（命名空间）为进程提供各自的系统资源视图，让同一台机器上的进程可以看到不同的进程列表、网络接口、主机名和文件系统挂载点。它是容器隔离的基础机制之一。

## 1. 基本概念

一个进程同时属于多种 namespace，每种负责一个方面。两个进程可以共享网络 namespace，却使用不同的挂载 namespace，不必将所有环境一起隔离。

例如，在容器内运行 `ps` 通常只能看到容器内的进程，而宿主机可以看到这些容器进程。同一个进程在不同的 PID namespace 中可以有不同的 PID：

```text
宿主机视角：PID 12345
容器内视角：PID 1
```

这仍是同一个进程，只是在不同层级的 PID namespace 中编号不同。PID namespace 具有层级关系，祖先 namespace 可以看到后代 namespace 中的进程，反向则不成立。

## 2. 常见类型

| 类型 | 隔离或虚拟化的内容 |
| --- | --- |
| PID | 进程 ID 与进程可见范围。 |
| Network | 网络接口、IP 地址、路由、端口等。 |
| Mount | 文件系统挂载点及挂载视图。 |
| UTS | 主机名与 NIS 域名。 |
| IPC | System V IPC、POSIX 消息队列等。 |
| User | 用户和组 ID 映射，以及相关权限范围。 |
| Cgroup | 进程看到的 cgroup 路径视图，不直接设置资源额度。 |
| Time | 部分时钟的时间偏移，例如单调时钟和启动时钟。 |

独立的 Mount namespace 不会自动复制文件内容；不同挂载视图仍可能访问同一份底层文件。独立的 Network namespace 也不会自动具备对外连接，需要另外配置接口、路由等。

## 3. 命令行使用

以下示例在安装了 util-linux 工具的 Linux 测试环境执行。`unshare` 创建新的 namespace，`nsenter` 加入已有的 namespace。所需权限取决于 namespace 类型、用户身份及系统策略。

### 3.1 创建独立的主机名环境

```bash
sudo unshare --uts /bin/bash
```

在新 shell 中执行：

```bash
hostname namespace-demo
hostname
```

输出为 `namespace-demo`。在宿主机另一个终端执行 `hostname`，仍然显示原来的主机名。执行 `exit` 退出示例 shell。

### 3.2 创建独立的进程视图

```bash
sudo unshare --pid --fork --mount-proc /bin/bash
```

- `--pid`：创建新的 PID namespace。
- `--fork`：创建子进程，让新 shell 进入这个 PID namespace。
- `--mount-proc`：挂载与新 PID namespace 对应的 `/proc`，同时创建 Mount namespace，避免影响宿主机挂载视图。

进入后执行：

```bash
echo $$
ps -ef
```

shell 在新 namespace 内的 PID 为 `1`，`ps` 通常只显示该环境内的进程；宿主机仍能看到它们。只创建 PID namespace 而沿用原来的 `/proc`，可能使 `ps` 展示的视图不符合预期。

PID 1 有特殊的信号处理和孤儿进程回收职责。这里的 shell 用于演示，长期运行的容器环境通常需要合适的初始化或进程回收方案；该 PID namespace 的初始化进程退出时，内核会终止其中其余进程。

### 3.3 加入已有的 namespace

假设目标进程在当前环境可见的 PID 为 `12345`：

```bash
# 与目标进程共享主机名环境
sudo nsenter --target 12345 --uts /bin/bash

# 与目标进程共享网络环境
sudo nsenter --target 12345 --net /bin/bash

# 加入目标进程的挂载和 PID namespace
sudo nsenter --target 12345 --mount --pid /bin/bash
```

只指定需要共享的类型即可。进入 Network namespace 不会同时切换挂载视图，例如 DNS 解析使用的 `/etc/resolv.conf` 仍取决于当前文件系统环境。

### 3.4 查看进程的 namespace

```bash
# 查看当前进程关联的 namespace
ls -l /proc/self/ns

# 对比两个进程的网络 namespace
readlink /proc/12345/ns/net
readlink /proc/67890/ns/net
```

对于同一种类型，如果两条链接显示相同的 namespace 标识，例如 `net:[4026531992]`，说明这两个进程共享该 namespace。读取其他进程的信息仍受权限检查约束。

## 4. 程序接口与子进程继承

| 接口 | 用途 |
| --- | --- |
| `clone` / `clone3` | 创建进程时，通过相应标志创建新的 namespace。 |
| `unshare` | 解除调用进程对部分系统环境的共享，创建新的 namespace。 |
| `setns` | 通过 namespace 文件描述符加入已有的 namespace，具体行为取决于类型。 |

创建子进程时，默认继承父进程的 namespace；之后的后代同样继承，不需要逐个加入。

PID namespace 是需要特别区分的情况：`unshare(CLONE_NEWPID)` 不改变调用者自身所属的 PID namespace，而是让后续创建的子进程进入新 namespace；`setns` 切换 PID namespace 时也影响后续创建的子进程。这就是命令示例使用 `--fork` 的原因。`nsenter --pid` 默认也会先创建子进程再执行目标命令。

进程通过 `exec` 替换程序时，通常保留已有的 namespace 归属。

## 5. 与 cgroup、容器的关系

| 机制 | 主要作用 |
| --- | --- |
| Namespace | 决定进程看到和共享哪套系统资源环境。 |
| Cgroup | 管理一组进程使用的资源额度、权重与统计。 |
| 容器运行时 | 组合 namespace、cgroup、根文件系统与权限控制等机制，构建和管理容器。 |

例如，Docker 的 `--network=host` 让容器共享宿主机的网络 namespace，而 CPU、内存限制仍由 cgroup 等机制实现。

单独创建某一种 namespace 只隔离对应的资源视图，不会自动形成完整的容器安全边界。尤其 User namespace 内的 UID 0 不等于拥有宿主机的全部 root 权限，其权限需要结合 ID 映射和所属的 user namespace 理解。
