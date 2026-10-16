# Agent Sandbox

Agent Sandbox 是 Kubernetes SIG Apps 下的沙箱管理项目，通过 CRD 和 Controller 管理隔离、有状态、具有稳定身份的执行环境，适用于 AI Agent、代码执行和开发环境等场景。

## 1. 项目定位

Agent Sandbox 负责执行环境的创建、分配和生命周期管理。底层隔离由容器运行时提供，可以通过 Kubernetes `RuntimeClass` 配合 gVisor 或 Kata Containers 使用。实际隔离能力取决于运行时与权限、网络等配置。

`kubernetes-sigs` 是 Kubernetes 社区 SIG 项目的 GitHub 组织；SIG 是 Special Interest Group（特别兴趣小组）的缩写。Agent Sandbox 是需要单独安装的生态项目。

## 2. 核心组件

| 组件 | 职责 |
|---|---|
| [Sandbox](Sandbox.md) | 声明一个具有稳定身份、可配置持久存储的执行环境；详见创建、连接与命令执行机制 |
| `SandboxTemplate` | 定义可复用的执行环境配置 |
| `SandboxWarmPool` | 维护预热环境池，缩短环境分配等待时间 |
| `SandboxClaim` | 申请执行环境，可从预热池领取 Sandbox |
| [Controller](Controller.md) | 根据自定义资源管理底层 Pod 等资源的生命周期；多个调谐循环运行在同一 Controller 进程中 |

## 3. Agent 与 Sandbox 的运行关系

### 3.1 Agent 在外部调用 Sandbox

Agent 主循环调用模型并决定下一步操作，通过 SDK 将命令或文件操作发送到 Sandbox。Shell、Python、浏览器或测试进程在 Sandbox 内执行，结果返回给 Agent。

```text
Agent 主循环 → Sandbox SDK → Sandbox 内的工具与进程
     ↑                              │
     └────────── 执行结果 ──────────┘
```

官方 ADK 示例将 `execute_python` 注册为 Agent 工具，由该工具创建 Sandbox、写入 Python 文件、执行代码并返回输出。

### 3.2 Agent 本身运行在 Sandbox 内

也可以将 Agent 程序部署到 Sandbox 内，由 Sandbox 管理其进程、依赖和持久化数据。模型可以通过外部 API 调用，也可以按部署方案在本地推理。

官方 OpenClaw 示例将持续运行的 Agent 放在 Sandbox 内；LangGraph 示例则在主容器中运行编程 Agent 和本地模型推理。

Agent Sandbox 支持这两种集成方式，具体选择取决于需要隔离和管理的范围。

## 4. 参考资料

- [GitHub 仓库](https://github.com/kubernetes-sigs/agent-sandbox)
- [官方文档](https://agent-sandbox.sigs.k8s.io/docs/)
- [ADK：将 Sandbox 作为 Agent 工具](https://agent-sandbox.sigs.k8s.io/docs/use-cases/examples/code-interpreter-agent-on-adk/)
- [OpenClaw：Agent 运行在 Sandbox 内](https://agent-sandbox.sigs.k8s.io/docs/use-cases/openclaw/)
- [LangGraph 编程 Agent 示例](https://agent-sandbox.sigs.k8s.io/docs/use-cases/examples/langchain/)
