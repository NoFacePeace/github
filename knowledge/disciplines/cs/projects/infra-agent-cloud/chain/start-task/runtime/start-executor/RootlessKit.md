# RootlessKit 网络环境

## 1. 工具分工

RootlessKit 帮助普通用户建立基于 Linux namespace 的隔离环境，常用于 rootless 容器。用户命名空间内部的 root 身份受宿主 UID/GID 映射约束，不等于宿主 root。

以下以 RootlessKit、slirp4netns、防火墙和 bwrap 组合启动任务为例：

| 工具 | 作用 |
| --- | --- |
| RootlessKit | 创建外层用户、挂载及所选网络命名空间，启动其中的目标命令，管理网络 helper。 |
| slirp4netns | 为非特权网络命名空间提供用户态网络出口。 |
| iptables / ip6tables | 在相应网络命名空间中设置访问规则。 |
| bwrap | 进一步构建任务的文件系统、进程和用户隔离，加载 seccomp。 |

可概括为：RootlessKit 建立环境，slirp4netns 提供出口，防火墙限制访问范围。使用 RootlessKit 本身不代表已完成文件、资源或系统调用限制。

## 2. 启动命令与五个选项

```sh
rootlesskit \
  --net=slirp4netns \
  --disable-host-loopback \
  --copy-up=/etc \
  --port-driver=none \
  --state-dir="$network_dir/rootlesskit" \
  /bin/sh /absolute/path/to/network-entrypoint.sh
```

这是通用命令结构，需提前准备状态目录的父目录和脚本，并替换实际路径；本文未在 Linux 环境执行示例。

| 选项 | 含义 |
| --- | --- |
| `--net=slirp4netns` | 选择 slirp4netns 网络驱动，由 RootlessKit 启动并配置 helper。 |
| `--disable-host-loopback` | 禁止通过 slirp 的宿主 loopback 通道访问宿主服务；不等于阻断所有宿主地址。 |
| `--copy-up=/etc` | 为命名空间中的 `/etc` 准备可写副本视图，供网络等配置使用，避免直接更改宿主目录内容；不是完整文件隔离。 |
| `--port-driver=none` | 不启用 RootlessKit 的入站端口转发驱动。 |
| `--state-dir=...` | 指定状态目录，用于本次实例的运行状态和管理信息，应避免并发实例共用。 |

这五个选项是一种集成配置，不是 RootlessKit 的全部可用选项，也不是每次使用都必须全部指定。

## 3. 选项与被执行命令的边界

整体语法是：

```text
rootlesskit [自身选项] 命令 [命令参数...]
```

从上例的 `/bin/sh` 开始，属于 RootlessKit 要在隔离环境中启动的子命令。防火墙脚本、FD 打开操作和 bwrap 启动参数不属于 RootlessKit 自身选项。

如果使用 `/bin/sh -c '脚本内容' 脚本标签 参数...`，脚本标签对应 Shell 的 `$0`，之后才依次是 `$1`、`$2` 等位置参数。脚本中的 `--` 可以用作自定义参数分隔符，不能据此认为它是 RootlessKit 选项。

## 4. Runner、子进程与 exec

```text
Runner：准备启动文件、资源分组和命令参数
  └─ 创建子进程并加入执行 cgroup
       └─ exec RootlessKit
            ├─ 管理 slirp4netns 网络 helper
            └─ 在隔离环境中启动网络脚本
                 ├─ 设置防火墙规则
                 ├─ 打开进程信息输出 FD
                 └─ exec seccomp 包装脚本
                      ├─ 打开过滤器输入 FD
                      └─ exec bwrap
                           └─ 启动目标程序
```

这是启动控制链。RootlessKit 的隔离环境中运行的是子进程，不是将脚本放入 RootlessKit 自己的进程中解释执行。`exec` 替换当前进程中的程序，保留所在 namespace 等运行上下文；每个箭头不一定增加一个常驻进程。

若 bwrap 复用外层网络命名空间，再创建内层用户命名空间并移除 capabilities，任务可使用已配置的网络，同时不具备修改外层网络规则所需的权限。该行为依赖实际参数，不能仅凭使用了 bwrap 就推断成立。

## 5. slirp4netns 如何提供出口

独立网络命名空间不能直接复用宿主网络接口。slirp4netns 接入隔离网络，在用户态处理流量，再通过宿主网络访问外部服务，不要求在宿主上配置特权网桥或内核 NAT。

```text
任务程序 → 隔离环境中的虚拟网卡 → slirp4netns → 宿主网络 → 外部服务
```

指定 `--net=slirp4netns` 后，由 RootlessKit 管理其启动和配置，不需要再手工启动同一实例的 slirp4netns。外部服务既可以是互联网，也可以是宿主网络能够到达的内网地址。

网络出口不等于访问授权。阻断宿主 IPv4 地址、限制 IPv6、实施域名白名单等，需要额外的网络规则或代理策略。独立 localhost 也不等于禁止通过其他 IP 访问宿主。

## 6. 网络规则示例与访问范围

以下是一种以限制宿主访问为目标的网络脚本策略，不是 RootlessKit 自动添加的默认规则。启动管理程序先在宿主侧通过 `ip -j -4 address show` 枚举 IPv4 地址，包括 DOWN 接口和网桥上的地址，再将地址列表传给隔离环境中的脚本。

脚本在 RootlessKit 创建的网络命名空间中执行，核心规则为：

```sh
# 阻断非 loopback 的 IPv6 出站流量
ip6tables -w -A OUTPUT ! -o lo -j REJECT

# 阻断经非 loopback 接口访问 0.0.0.0/8
iptables -w -A OUTPUT ! -o lo -d 0.0.0.0/8 -j REJECT

# 对启动时枚举到的每个宿主 IPv4 地址重复执行
iptables -w -A OUTPUT ! -o lo -d "$host_address" -j REJECT
```

`OUTPUT` 检查当前网络命名空间发出的流量；`! -o lo` 表示出口接口不是 loopback；`-d` 指定目标地址或网段；`REJECT` 表示拒绝匹配流量。这里修改的是隔离网络中的规则，不是宿主网络命名空间的防火墙。

### 6.1 允许与拒绝范围

| 流量或能力 | 示例配置的处理 |
| --- | --- |
| 沙箱内部的 `127.0.0.1`、`::1` 通信 | 不被上述规则拒绝，可用于同一网络命名空间内的服务通信。 |
| 经非 loopback 接口访问启动时识别的宿主 IPv4 地址 | 拒绝，包括已枚举的网桥地址。 |
| 经非 loopback 接口访问 `0.0.0.0/8` | 拒绝。 |
| 非 loopback 的 IPv6 出站流量 | 拒绝。 |
| 通过 slirp 的宿主 loopback 通道访问服务 | 由 `--disable-host-loopback` 禁止。 |
| 其他 IPv4 内网或互联网地址 | 上述规则未禁止，仍需路由、slirp4netns 和目标服务支持。 |
| 从外部主动连接沙箱服务 | `--port-driver=none` 不配置入站端口转发。 |
| 沙箱主动请求的响应 | 不因关闭入站端口转发而被禁止，仍可作为已建立通信的一部分返回。 |

这组规则没有域名白名单、IPv4 出站端口白名单，也没有整体封锁内网。未被这些规则禁止不代表连接必然成功，还可能受其他防火墙、路由和服务配置限制。

### 6.2 宿主访问限制的边界

- 地址列表是启动时的快照。宿主新增或变更 IPv4 地址后，旧规则不自动包含新地址，应刷新规则或重建执行环境。
- 通过外部负载均衡器、NAT 或代理绕回宿主的流量，其目标地址可能不在宿主本地地址列表中，不能依靠这组规则保证阻断。
- `--disable-host-loopback` 只限制对应的宿主 loopback 通道，不代替宿主地址阻断规则，也不等于封锁所有访问宿主的路径。
- 只读 hosts 文件影响名称解析，程序仍可能直接访问 IP；访问边界应由网络规则等机制实施。

因此，应将此示例理解为“限制已识别的直接宿主访问路径”，不能直接表述为“绝对无法访问宿主机”。

## 7. 依赖和运行条件

- 在 Linux 环境安装 RootlessKit、slirp4netns，并确保启动时的 `PATH` 能找到它们。
- 系统和外层容器策略需允许所需的非特权 user/network namespace。
- 按所用 UID/GID 映射方式准备 `newuidmap`、`newgidmap`，以及 `/etc/subuid`、`/etc/subgid` 等配置。
- 若脚本设置防火墙，需提供 iptables、ip6tables 及相应内核支持；枚举宿主地址通常还使用 iproute2 的 `ip`。
- 状态目录和锁文件应位于运行用户可写的私有目录。执行结束后由启动管理程序停止 helper、回收进程并清理状态。

工具安装成功不代表 namespace、映射或外层安全策略满足要求，应在目标 Linux 主机验证。可用 `rootlesskit --help` 和 `slirp4netns --help` 核对安装版本支持的选项。

## 8. 官方资料

- [RootlessKit 官方仓库](https://github.com/rootless-containers/rootlesskit)：用途、依赖及使用入口。
- [RootlessKit 网络驱动说明](https://github.com/rootless-containers/rootlesskit/blob/master/docs/network.md)：网络驱动与相关配置。
- [slirp4netns 官方仓库](https://github.com/rootless-containers/slirp4netns)：非特权网络命名空间的用户态联网原理和用法。
