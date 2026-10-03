# Sandbox

## 1. 开源沙箱

### 1.1 CubeSandbox

CubeSandbox 是腾讯云开源的 AI Agent 沙箱服务，用于在隔离环境中执行 AI 生成的代码、Shell 命令和自动化任务。它基于 RustVMM 和 KVM，为每个沙箱提供独立的 MicroVM 和操作系统内核，适用于代码执行、数据分析和浏览器自动化等场景。

主要特点：

- **快速创建**：通过资源池和快照克隆加速沙箱创建。
- **状态管理**：支持快照、克隆、回滚，以及暂停和恢复，便于 Agent 重试或探索不同方案。
- **网络控制**：支持沙箱间网络隔离、出站访问过滤和凭证代理。
- **部署与集成**：支持单机和多节点集群部署，提供 E2B SDK 兼容接口。

仓库链接：[TencentCloud/CubeSandbox](https://github.com/TencentCloud/CubeSandbox)

### 1.2 E2B

E2B 是面向 AI Agent 的开源沙箱平台，让 Agent 通过 API 或 SDK 创建隔离环境，执行 AI 生成的代码、运行命令和操作文件。它基于 Firecracker MicroVM，适用于代码执行、数据分析和自动化任务等场景。例如，Agent 可以将生成的 Python 脚本放入沙箱执行，再取回输出、文件和图表。

主要特点：

- **隔离执行环境**：通过 MicroVM 为代码和工具提供隔离的运行环境。
- **SDK 支持**：提供 Python 和 JavaScript/TypeScript SDK，用于创建沙箱、执行命令和读写文件。
- **自定义环境**：支持通过模板预装依赖和工具。
- **部署方式**：提供 E2B Cloud 托管服务，也可以基于开源项目自行部署。

CubeSandbox 提供 E2B SDK 兼容接口，便于已有 E2B 客户端接入；具体功能的兼容程度需按版本确认。

仓库链接：[e2b-dev/E2B](https://github.com/e2b-dev/E2B)

官网：[e2b.dev](https://e2b.dev/)

### 1.3 OpenKruise Agents

OpenKruise Agents 是 OpenKruise 的子项目，用于在 Kubernetes 上管理 AI Agent 沙箱和工作负载。它提供沙箱编排与生命周期管理能力，在集群中创建、分配、暂停、恢复和回收 Agent 执行环境，适用于 Agent 工具执行、云端开发工作空间和强化学习等场景。

主要特点：

- **资源管理**：通过资源池和动态调整资源，加快沙箱分配并降低运行成本。
- **状态保存**：提供休眠和检查点能力，保存内存、文件系统等运行状态；具体支持取决于运行环境。
- **访问管理**：管理用户身份、会话，以及请求到对应沙箱的路由。
- **API 与 SDK**：提供 Kubernetes CRD API，支持通过 E2B SDK 创建和管理沙箱。

相比 CubeSandbox 和 E2B，OpenKruise Agents 更侧重 Kubernetes 上的沙箱编排与生命周期管理。其隔离强度取决于底层容器或沙箱运行时，使用该项目本身不意味着获得 MicroVM 级隔离。

仓库链接：[openkruise/agents](https://github.com/openkruise/agents)

官方文档：[OpenKruise Agents](https://openkruise.io/kruiseagents/introduction)
