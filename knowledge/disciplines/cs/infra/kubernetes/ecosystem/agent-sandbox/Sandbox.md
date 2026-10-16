# Sandbox

Sandbox 是 Agent Sandbox 项目提供的 Kubernetes 自定义资源，用于声明和管理一个具有稳定身份、可配置持久存储的执行环境。Controller 根据资源配置管理底层 Pod 等对象。

## 1. 执行环境与执行服务

Sandbox 管理执行环境；容器镜像与启动配置决定环境内运行什么程序。

| 层次 | 职责 |
|---|---|
| Sandbox CRD 与 Controller | 声明环境、创建 Pod、管理生命周期 |
| 容器镜像 | 提供执行服务、Python、Shell、依赖等程序和文件 |
| 容器启动配置 `command` / `args` | 决定容器启动时运行哪个程序 |
| 执行服务 | 接收远程请求、启动命令进程、返回结果 |
| 隔离运行时 | 提供底层隔离，例如 gVisor 或 Kata Containers |

Agent Sandbox 不会自动为任意镜像注入执行服务。镜像可以运行 Python Runtime、`sandboxd` 或普通应用；SDK 的命令与文件接口需要与实际执行服务匹配。

## 2. 创建与定位

可以直接创建 `Sandbox` 资源，也可以通过 `SandboxClaim` 申请环境，结合模板和预热池完成分配。

官方 Python SDK 的创建示例：

```python
from k8s_agent_sandbox import SandboxClient

client = SandboxClient()
sandbox = client.create_sandbox(
    warmpool="python-sandbox-warmpool",
    namespace="default",
)

try:
    result = sandbox.commands.run("python3 run.py")
    print(result.stdout)
finally:
    sandbox.terminate()
```

这个示例要求预热池已配置相应运行时镜像，且沙箱内存在 `run.py`。返回的 `sandbox` 对象代表分配的目标环境，后续操作通过它发送。

跨进程或跨会话重新连接时，需要保存 Sandbox 标识、Namespace，以及多集群场景下的集群信息。保存标识不会延长环境生命周期，需要另外配置清理策略。连接已有环境可使用 SDK 的 `get_sandbox`，具体参数以所安装版本为准。

## 3. 通信路径

### 3.1 管理请求

创建、查询和删除等操作通过 Kubernetes API 完成。SDK 需要集群连接配置与相应权限。Sandbox 本身不需要模型 API Key；Agent 调用云端模型时才需要对应凭据。

### 3.2 执行请求

通过 Router 访问时：

```text
Agent / SDK → Gateway（可选）→ Sandbox Router → 沙箱内执行服务
```

Router 接收并转发请求；命令最终由沙箱内的执行服务处理。

Python SDK 的集群内模式可以绕过 Router，直接连接 Pod IP 或集群内 DNS。SDK 通过 Kubernetes 查询目标地址，调用者通常不必手动提供 IP，但必须具备查询权限和集群网络可达性。本地开发还可以通过 `kubectl port-forward` 连接 Router。

### 3.3 SDK 与协议

SDK 是调用封装，不是一种协议，也不能自动兼容任意执行服务。

| 操作或运行时 | 通信方式 |
|---|---|
| Kubernetes 资源管理 | Kubernetes API，通常通过 HTTPS |
| Python Runtime 示例 | HTTP API，例如 `/execute` |
| `sandboxd` 命令执行 | gRPC `ProcessService` |

SDK 针对支持的运行时实现相应调用逻辑。自定义执行服务需要实现客户端支持的接口，或增加客户端适配。接口名称相似不代表 E2B SDK/API 兼容，不能假定只替换服务地址即可使用。

## 4. 命令执行

### 4.1 谁提供命令

具体命令由调用者或 Agent 提供，例如 `python3 run.py`、`pytest`、`ls -la`。SDK 传递命令，执行服务启动相应进程；镜像中必须安装对应程序，执行所需文件也必须已存在。

通常无需为每条命令编写独立 API，执行服务提供通用的启动进程能力。

### 4.2 Python Runtime 的实现

官方 `examples/python-runtime-sandbox` 使用 Python 和 FastAPI 实现一个常驻执行服务。它与 `sandboxd` 是不同实现。

```text
sandbox.commands.run("python3 run.py")
                 ↓
HTTP POST /execute，携带 command
                 ↓
FastAPI 接收请求，shlex.split 解析参数
                 ↓
subprocess.Popen 启动 Python 子进程
                 ↓
收集 stdout、stderr、退出码并返回
```

其核心逻辑可简化为：

```python
args = shlex.split(command)
process = subprocess.Popen(
    args,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True,
    cwd=base_dir,
    start_new_session=True,
)
stdout, stderr = process.communicate(timeout=timeout)
```

实际服务还处理超时，并在超时时终止进程组。执行服务在容器运行期间常驻，每条执行请求启动新的命令进程；容器停止或重启时，服务也会停止或重启。

该示例默认直接启动程序，不经过 Shell。使用管道、重定向等 Shell 语法时，需要显式调用 Shell，例如 `sh -c 'ls | grep py'`。

### 4.3 不使用执行服务的方式

| 方式 | 是否需要沙箱内的远程执行服务 |
|---|---|
| 容器启动配置执行 `python3 run.py` | 不需要，由容器运行时启动 |
| `kubectl exec ... -- python3 run.py` | 不需要，使用 Kubernetes exec 机制 |
| 沙箱内 Agent 使用 `subprocess` | 不需要，直接创建本地子进程 |
| SDK 调用 `sandbox.commands.run(...)` | 需要与 SDK 匹配的执行服务 |

执行服务的作用是让外部客户端远程启动程序并获取结果；单纯在容器内运行程序不要求额外执行服务。

## 5. 参考资料

- [Sandbox 项目概览](https://github.com/kubernetes-sigs/agent-sandbox)
- [Python SDK 说明](https://github.com/kubernetes-sigs/agent-sandbox/blob/main/clients/python/agentic-sandbox-client/README.md)
- [Python Runtime 示例](https://github.com/kubernetes-sigs/agent-sandbox/tree/main/examples/python-runtime-sandbox)
- [Python Runtime 执行服务源码](https://github.com/kubernetes-sigs/agent-sandbox/blob/main/examples/python-runtime-sandbox/main.py)
- [客户端依赖与 sandboxd gRPC 说明](https://github.com/kubernetes-sigs/agent-sandbox/blob/main/clients/python/agentic-sandbox-client/pyproject.toml)
