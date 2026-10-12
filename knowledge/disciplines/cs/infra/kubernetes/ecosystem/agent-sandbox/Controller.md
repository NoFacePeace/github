# Controller

Agent Sandbox 的 Controller 服务负责监听 Kubernetes 自定义资源，将资源声明转换为实际环境，并持续调谐资源状态。它负责环境管理；沙箱内的命令执行由执行服务处理。

## 1. 实际部署结构

根据当前上游 `main` 分支的启动代码和部署清单，多个 Controller 注册到同一个 controller-runtime Manager，运行在同一个 `agent-sandbox-controller` 进程中。

```text
agent-sandbox-controller Deployment
└── Pod / controller 容器
    └── agent-sandbox-controller 进程
        └── Manager
            ├── Sandbox Controller
            ├── SandboxClaim Controller
            ├── SandboxTemplate Controller
            └── SandboxWarmPool Controller
```

部署清单默认配置一个副本，并启用 Leader Election。Controller 的逻辑数量不等于 Deployment 或进程数量；一个进程可以运行多个调谐循环。具体发布版本应以其启动代码和安装清单为准。

## 2. 各 Controller 的职责

| Controller | 监听资源 | 职责 |
|---|---|---|
| Sandbox | `Sandbox` | 管理底层 Pod 等资源及环境生命周期 |
| SandboxClaim | `SandboxClaim` | 处理环境申请与 Sandbox 分配，更新申请状态 |
| SandboxTemplate | `SandboxTemplate` | 处理可复用环境模板的调谐逻辑 |
| SandboxWarmPool | `SandboxWarmPool` | 管理预热 Sandbox 池，按配置维护池中环境 |

每个 Controller 有自己的 Reconciler 和调谐逻辑。它们通过 Kubernetes API 中的资源和状态协作，不代表一个请求必须依次同步调用所有 Controller。

## 3. 扩展启用方式

启动参数 `--extensions` 控制是否注册扩展 Controller，代码中的默认值为 `false`。

| 启动配置 | 注册的 Controller |
|---|---|
| 未启用 `--extensions` | Sandbox |
| 启用 `--extensions` | Sandbox、SandboxClaim、SandboxTemplate、SandboxWarmPool |

启动代码先创建 Manager 并注册 `SandboxReconciler`，再在 `if extensions` 分支中注册三个扩展 Reconciler，最后通过 `mgr.Start(...)` 启动。

启用扩展功能还需要安装相应 CRD 并配置访问权限；仅修改启动参数不能代替这些资源的安装。

## 4. 创建环境的调谐路径

### 4.1 直接创建 Sandbox

```text
kubectl / Kubernetes 客户端
          ↓
Kubernetes API：创建 Sandbox CR
          ↓
Sandbox Controller 调谐
          ↓
创建并管理 Pod 等资源
```

### 4.2 通过 Claim 申请环境

使用 Python SDK 的预热池创建接口时，SDK 提交 `SandboxClaim`，由扩展 Controller 处理申请。

```text
SDK → Kubernetes API：创建 SandboxClaim CR
                             ↓
                    SandboxClaim Controller
                             ↓
                    分配预热池中的 Sandbox
                             ↓
                    更新 Claim 状态与就绪信息
                             ↓
                    SDK 等待就绪并返回环境对象
```

预热池中的 Sandbox 由 WarmPool Controller 维护，其底层 Pod 由 Sandbox Controller 管理。申请环境与执行命令属于不同链路；SDK 的命令请求最终交给沙箱内的执行服务。

详见 [Sandbox](Sandbox.md) 中的创建、连接与命令执行机制。

## 5. 参考资料

- [Controller 启动代码](https://github.com/kubernetes-sigs/agent-sandbox/blob/main/cmd/agent-sandbox-controller/main.go)
- [Controller Deployment 清单](https://github.com/kubernetes-sigs/agent-sandbox/blob/main/k8s/controller.yaml)
- [核心 Controller 源码目录](https://github.com/kubernetes-sigs/agent-sandbox/tree/main/controllers)
- [扩展 Controller 源码目录](https://github.com/kubernetes-sigs/agent-sandbox/tree/main/extensions/controllers)

[返回 Agent Sandbox](AgentSandbox.md)
