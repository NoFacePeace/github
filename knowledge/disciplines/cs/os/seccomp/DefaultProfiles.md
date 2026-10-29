# 容器默认 Seccomp 策略

## 1. 默认策略的来源

Linux 提供 seccomp 机制、BPF 指令格式和过滤动作，但没有为所有容器或沙箱规定一份统一的过滤器模式默认策略。Docker、containerd、Podman 等项目各自维护默认规则，在通用程序兼容性和减少内核攻击面之间取舍。

| 项目 | 策略入口 | 主要特点 |
| --- | --- | --- |
| Docker / Moby | [默认策略 JSON](https://github.com/moby/profiles/blob/main/seccomp/default.json) | 默认拒绝，通过允许规则开放常用调用，并根据架构、capabilities、调用参数等调整。 |
| containerd | [默认策略生成代码](https://github.com/containerd/containerd/blob/main/contrib/seccomp/seccomp_default.go) | 通过代码生成策略，包含平台和权限条件；实际是否加载取决于调用方和运行配置。 |
| Podman / containers-common | [策略 JSON](https://github.com/containers/common/blob/main/pkg/seccomp/seccomp.json) | 维护供容器工具使用的默认策略，包含允许、错误返回及条件规则。 |

“项目提供默认策略”不代表所有启动入口都会自动应用。例如，不能仅因底层使用 containerd，就判断 Kubernetes Pod 已加载其默认 seccomp 策略，需要核对实际配置。

## 2. 允许列表与拒绝列表

主流容器默认策略通常采用允许列表：

```text
系统调用及其参数满足允许条件 → 允许
匹配特定拒绝规则             → 返回指定错误等
其余                         → 执行默认拒绝动作
```

这与“默认允许，仅拒绝十几个调用”的精简拒绝列表不同：

| 情况 | 默认拒绝的允许列表 | 默认允许的拒绝列表 |
| --- | --- | --- |
| 未列出的系统调用 | 拒绝 | 允许 |
| 内核新增系统调用 | 通常需要更新规则后才能使用 | 通常自动放行 |
| 维护重点 | 覆盖程序所需调用和兼容路径 | 持续识别应禁止的接口 |

默认拒绝不表示所有拒绝都返回同一错误。策略可以针对不同调用返回 `EPERM` 或 `ENOSYS`，也可以只允许某些参数组合。

## 3. 常见限制及原因

以下是理解策略时可关注的类别，不是三个项目完全相同、永久不变的禁止清单。实际行为需结合策略版本、架构、capabilities 和参数条件判断。

| 类别 | 代表系统调用 | 限制考虑 |
| --- | --- | --- |
| 内核切换 | `kexec_load`、`kexec_file_load` | 普通应用不需要切换宿主内核。 |
| 内核模块 | `init_module`、`finit_module`、`delete_module` | 控制内核扩展，通常属于宿主维护。 |
| swap 管理 | `swapon`、`swapoff` | 操作系统级交换空间。 |
| BPF、性能观测、缺页处理 | `bpf`、`perf_event_open`、`userfaultfd` | 暴露额外内核接口，特殊任务需单独评估。 |
| 文件句柄访问 | `open_by_handle_at` | 涉及通过句柄访问文件及相关权限。 |
| 异步 IO | `io_uring_setup`、`io_uring_enter`、`io_uring_register` | 需要结合内核安全风险和业务回退能力评估。 |
| namespace、挂载、调试、keyring | `clone`、`unshare`、`mount`、`ptrace`、`keyctl` 等 | 项目间及条件间差异较大，不能一概视为默认禁止。 |

Docker 对部分重要限制及原因有[官方说明](https://docs.docker.com/engine/security/seccomp/)。containerd 2.0 的[变更说明](https://github.com/containerd/containerd/blob/main/docs/containerd-2.0.md#io_uring_-syscalls-are-disallowed-by-default)记录了默认禁用 io_uring 的决定。

即使某调用已被 capabilities 检查拒绝，seccomp 仍可显式限制它，形成额外保护。反过来，seccomp 允许调用，也不代表能绕过权限和 namespace 边界。

## 4. 如何阅读与复用

### 4.1 先看默认动作，再看条件

阅读 JSON 或生成代码时，依次检查：

1. `defaultAction`：没有规则匹配时如何处理。
2. `architectures` 或 `archMap`：支持哪些系统调用 ABI。
3. 每条规则的调用名称和 `action`。
4. `args`：是否只允许特定参数值或标志位。
5. `includes`、`excludes` 或生成逻辑：是否受 capabilities、内核版本等条件影响。
6. 特定错误返回：程序是否依赖 `ENOSYS` 才会尝试替代实现。

不能仅搜索某个系统调用名称是否出现在文件中，就判断它最终允许或拒绝。

### 4.2 接入自建沙箱

Docker 等工具读取的 JSON 是规则描述；Bubblewrap 的 `--seccomp FD` 接收经典 BPF 二进制指令，不能直接读取 Docker JSON。

复用时需要解析规则、处理架构和权限条件，再通过 libseccomp 等方式生成 BPF，或使用已有的兼容加载实现。只摘取部分禁止项属于自定义策略，不能称为完整复用了容器默认策略。

建议固定项目的发布版本或 commit，并记录实际加载配置。本文引用的 `main` 分支会变化，是资料入口，不是可复现的版本快照。

### 4.3 验证策略效果

验证需要同时覆盖：

- 过滤确实安装，代表性受限调用按预期被拒绝，后续子进程仍继承规则。
- 实际任务正常，包括语言运行时、依赖安装、构建、浏览器、调试和会话恢复等需要的路径。

单次 `EPERM` 可能来自权限检查，单独看到 `Seccomp: 2` 也不能证明加载的是预期策略；需要结合部署版本、规则内容和探针结果确认。

## 5. 资料范围

资料核对日期：2026-10-29。本文说明主流策略的结构与复用方法，不提供跨版本固定不变的完整系统调用清单。升级内核、容器引擎或应用依赖时，应重新检查兼容性及策略变化。
