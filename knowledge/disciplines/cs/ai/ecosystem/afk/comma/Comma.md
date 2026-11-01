# Comma

Comma 是 AFK 开发的开源个人 AI Agent，提供官方托管服务和自托管方式。它将较大的用户目标转为持久任务，通过 Agent 循环持续规划、执行、检查和验证，并在需要用户决策时等待。本文依据官方 README、产品模型和运行时文档整理，核对日期为 2026-11-01；功能描述为官方说明，不代表独立实测。

## 1. 产品定位

Comma 强调 Sessionless：聊天只是交互入口，任务才是持久的执行状态。工作、记忆、上下文和进行中的任务可以跨设备、聊天和界面延续。

| 层次 | 作用 |
|---|---|
| 用户目标 | 描述希望完成的工作 |
| Tasks | 保存执行状态，持续推进目标 |
| Agent 与模型 | 执行规划、操作、检查和验证 |
| 设备与环境 | 提供云端工作环境，以及用户电脑上的文件和操作能力 |

Comma 属于面向用户的个人助手与任务编排产品。Claude Agent SDK 是用于开发 Agent 应用的工具包，Claude Code 等工具则可作为 Comma 调用或注册的执行者。

## 2. 任务与交互能力

| 能力 | 官方说明 |
|---|---|
| 持久任务 | 简单问题直接回答，较大工作转换为 Task，并检查完成情况 |
| 任务看板 | 通过 `In progress`、`Needs Review`、`Done` 展示状态 |
| 人工决策 | 遇到需要用户判断的事项，任务进入 `Needs Review` |
| 多 Agent 协作 | 将工作分配给不同 Agent 和模型；可接入 Codex、Claude Code 等执行者 |
| 主动通知与例行工作 | 需要用户介入时发出消息，按设置时间提供简报 |
| 跨设备执行 | 结合云端环境与个人设备，延续任务和上下文 |

Salix 是 Comma 的多 Agent 运行时，也提供设备连接能力。连接 Mac 后可读取设备文件；开启 `Allow operations` 后，还可修改文件、运行命令和向其他编程 Agent 分派工作。实际能力取决于设备连接、权限和运行环境。

## 3. 后台 Loops

Comma 中的 Loop 是 Agent 编写的小程序，用于监控收件箱、代码仓库或信息源，仅在出现需要处理的情况时唤醒 Agent。官方说明安静时段不消耗大模型 token。

此处的 Loop 侧重后台监控与唤醒，与一次任务中“调用模型 → 执行工具 → 反馈结果”的 Agent Loop 含义不同。

## 4. 运行架构

### 4.1 自有 Agent 与外部 Agent

Comma 有自己的 Agent 执行体系。Salix 运行内部 Agent，管理模型调用、工具执行、上下文、等待状态及持久化恢复，同时允许将工作委托给外部编程 Agent。

| 角色 | 职责 |
|---|---|
| Router | 理解用户意图、协调工作、委派任务并管理任务状态 |
| Worker | 执行具体任务，使用模型和工具并发布结果消息 |
| 外部执行者 | 接入 Claude Code、Codex 等，利用其已有执行能力完成工作 |

即使不接入外部编程 Agent，内部 Agent 也可以调用配置的模型服务与工具运行任务。自有 Agent 指运行与控制逻辑，不意味着 Comma 自己训练了基础模型。

与常见聊天式 Agent 相比，其核心“调用模型 → 执行工具 → 反馈结果”循环没有根本变化；区别主要在持久任务、长期协调、多执行者接入、后台唤醒和故障恢复的产品整合。这些能力并非 Comma 独有，也不意味着模型判断更准确。

### 4.2 运行时可靠性

运行时主要使用 Elixir/OTP，关键系统组件使用 Lean，并使用 Lean 与 TLA+ 验证部分运行时属性及分布式协议。

官方描述的属性包括：故障重启后不静默丢失已接收的工作、失败回复不计为完成、记录有副作用的操作、避免中断的修改操作被意外重复执行，以及阻止过期 Worker 覆盖较新的状态。

形式化验证针对运行时规则和故障处理，不证明模型判断正确或任务必然成功。

## 5. 数据模型

以下是领域概念及其关系，不是数据库表名清单。

| 模型 | 含义与关系 |
|---|---|
| User / Membership | 用户身份及对工作范围的访问权 |
| Workspace / Salix Group | Workspace 同时承担产品授权与 Agent 工作范围，并对应当前 Salix Group |
| Agent | 保存身份、配置与运行环境绑定，角色包括 Router、Worker |
| Conversation | 保存可见协作消息；主要类型为 `user_chat` 和 `agent_task` |
| Message | Conversation 中有序保存的消息，记录具体发送者 |
| Participant | Conversation 的接收目标；Agent 目标明确绑定 `agent_id` 与 `session_id` |
| Task | 以 `agent_task` Conversation 表达的持久工作对象，普通任务指定一个负责的 Worker |
| Runtime Session | Agent 的执行上下文，包含模型历史、工具结果与运行控制状态 |
| Artifact | 任务产生的工作结果 |
| Background Loop | 后台监控或周期执行对象，可与 Task 关联 |

任务状态以 Conversation 的 `status` 为权威，界面卡片展示其投影。官方当前任务文档还明确区分普通任务、设备执行任务和 Loop Task，具体执行及结束规则因任务类型而异。

## 6. Sessionless 与任务延续

### 6.1 仍然保留会话

Sessionless 主要描述用户不必管理会话的产品体验。底层仍有 Conversation 和 Runtime Session：前者保存用户可见的协作事实，后者保存 Agent 的执行上下文。

每个已配置的当前 Router 有一个 canonical Runtime Session；Worker 可以为不同任务使用独立 Session。Comma 当前就绪 Group 共用一条固定的 Router Conversation，因此不能将“跨聊天”理解为产品一定为每次交流创建新 Chat。

普通模型输出不会自动写成可见 Message，内部可见回复需要通过显式消息工具发送。Runtime transcript 与用户看到的消息历史是不同的数据边界。

### 6.2 持久任务如何继续

以“跟踪 PR #12，处理新评论”为例：

1. Router 创建具有稳定身份的 Task，记录要求并指定 Worker。
2. Task 保留消息历史和状态，Worker Session 保存执行上下文。
3. 对于持续监控任务，关联的 Background Loop 在出现事件时唤醒目标 Session。
4. 后续指令投递给同一个 Task；关闭前端窗口不删除服务端任务。
5. 运行进程重启后可重新加载持久状态；已接收输入和工具结果按运行时规则保留，中断的修改操作不会直接自动重做。

核心是 Task 独立于某次聊天交互持久存在。“保存任务”与“继续执行”是不同能力；后者还依赖运行环境、事件唤醒、权限及模型服务。持久化也不意味着所有历史永久保留在模型上下文中，运行时仍可能压缩上下文。

### 6.3 谁定位目标 Task

| 输入入口 | 目标如何确定 |
|---|---|
| 任务页面或已绑定的任务入口 | 应用已持有任务标识，消息按明确目标路由，无需模型猜测 |
| 主聊天中自然语言指代旧任务 | Router 负责理解意图并结合可用任务信息选择目标；匹配不明确时应澄清 |

在后一种交互中，语义判断由 Router 的模型承担，服务端负责权限校验与按目标标识投递，Worker 执行任务。官方任务文档要求后续输入使用预期 Task，而非直接选择最近活跃的任务。

具体任务查询工具、搜索索引、候选排序和消歧算法尚未逐项核对源码，不能据此推断 Comma 使用了固定的自动匹配算法，也不能假设任意新入口都会自动找到正确任务。

## 7. 开源与部署

- 官方托管入口为 `comma.surf`，同时提供 Web 和 Mac 客户端相关代码。
- 原创代码使用 `AGPL-3.0-only` 许可，第三方组件保留各自许可。
- 支持通过 Docker Compose 构建单节点实例，依赖 PostgreSQL、Redis、MinIO、ClickHouse 等服务。
- 真实模型调用需要配置相应模型服务与凭据，可能产生模型服务费用。
- 自托管仓库未包含部分会议及 Agent VMM 宿主所需的运行时；部署前应核对所需功能的依赖。

## 8. 参考资料

- [Comma 官网](https://comma.surf/)
- [官方仓库与 README](https://github.com/AFK-surf/Comma)
- [许可证](https://github.com/AFK-surf/Comma/blob/main/LICENSE)
- [自托管部署说明](https://github.com/AFK-surf/Comma/blob/main/systems/DEPLOYMENT.md#self-hosted-compose)
- [验证说明](https://github.com/AFK-surf/Comma/blob/main/docs/verification.md)
- [核心产品模型](https://github.com/AFK-surf/Comma/blob/main/docs/architecture/PRODUCT_MODEL.md)
- [领域概念清单](https://github.com/AFK-surf/Comma/blob/main/docs/architecture/DOMAIN_CONCEPTS.md)
- [系统架构](https://github.com/AFK-surf/Comma/blob/main/systems/README.md)
- [Agent 运行时](https://github.com/AFK-surf/Comma/blob/main/docs/agent-runtime.md)
- [任务与后台执行](https://github.com/AFK-surf/Comma/blob/main/docs/salix/tasks-background-execution.md)
