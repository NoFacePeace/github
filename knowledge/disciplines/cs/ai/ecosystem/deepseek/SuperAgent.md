# Super-agent

Super-agent 是 DeepSeek Harness 团队探索的自主 Agent 系统设计，强调持续接收消息与环境事件、管理多项工作，并减少对人类注意力的占用。官方公开了 `dsh-superagent-prototype` 研究原型。它不是新的基础模型，原型也不是正式 DeepSeek 产品。本文依据官方仓库 README 与实现架构文档整理，核对日期为 2026-11-01。

## 1. 架构定位

Super-agent 将协调与具体执行分开：协调层判断哪些消息需要行动、如何关联已有工作，执行层在独立上下文中完成任务。基础的模型与工具循环仍然存在，系统在其上增加事件接入、持久工作管理、路由、授权和人工介入机制。

| 层次 | 主要职责 |
|---|---|
| Router / 协调层 | 接收消息与事件，判断是否需要行动，决定创建、关联或停止工作 |
| Worker / 执行层 | 在独立 Session 和目录中执行具体工作，报告结果或请求人工介入 |

公开实现使用主 Session、Assessor、Executor 和 Engine 等组件承载这些职责，不应将概念中的 Router / Worker 直接等同于某个同名类。Engine 是领域对象的唯一写入入口，模型决策通过工具和领域命令落地，并受策略校验约束。

## 2. 公开原型的范围

`dsh-superagent` 以 [DeepSeek Harness（DSH）](DSH.md) 插件包的形式运行，主要面向单个 GitHub 仓库的主动编程协作。

公开原型是可运行的部分实现。“DSH 插件包”描述其集成与运行形式：原型实现事件接入、持久工作项管理、路由和执行协调，底层模型调用与工具运行能力依赖 DSH。完成环境、账号与权限配置后，可以运行这些已实现的流程；这不代表设计文章中的全部能力都已实现，也不代表原型已达到生产可用水平。

- 监听 GitHub，可选接入飞书。
- 在获授权人员请求工作，或团队规则允许行动时处理工作。
- 创建持久工作项，在独立 Agent Session 和目录中执行。
- 将结果回复到原始讨论中，例如给出 Pull Request。
- 通过 Web 面板展示需要人员处理的事项。

```text
GitHub / 飞书消息与事件
          ↓
协调层评估：新工作、已有工作的补充、停止或忽略
          ↓
Ticket → Assignment → Executor / Worker 执行
          ↓
结果、PR 或人工问题 → 原始讨论 / 管理面板
```

“主动”不意味着任意事件都能授权执行。原型强调仅在被请求或明确规则允许时行动，并在有人需要处理事项或可以使用结果时发言。

## 3. 主要领域对象与实现组件

| 对象或组件 | 用途 |
|---|---|
| Trigger | 来源插件接入的消息或事件，进行去重与路由 |
| Thread | 外部讨论与关联工作的边界，例如 Issue、PR 讨论 |
| Eval / Decision | 评估候选事件，并记录接受、停止、忽略或不可用等决策 |
| Ticket | 持久工作项，保存工作关联、状态和后续输入 |
| Assignment | 一次具体执行委派，记录启动、输入交付及执行报告 |
| Engine | 通过事务集中修改领域对象，校验命令与状态转换 |
| Outbox / Effects | 保存待发送或待执行的外部工作，在事务提交后处理 |

已有工作的后续消息可以按订阅关系送到 Ticket；主 Session 也可通过关联命令将消息路由到指定 Ticket。因此，任务延续依赖稳定工作标识与路由关系，并非每次消息都重新创建工作。

## 4. 与 Comma 的比较

两者都采用协调与执行分离、持久工作对象、独立执行上下文，以及事件驱动的工作方式。

| 方面 | Super-agent 公开原型 | Comma |
|---|---|---|
| 当前定位 | 单仓库编程协作的研究原型 | 面向用户的个人助手与任务平台 |
| 持久工作对象 | Ticket | Task，以 `agent_task` Conversation 表达 |
| 执行基础 | DeepSeek Harness 插件包 | 自有 Salix 多 Agent 运行时，也能接入外部执行者 |
| 主要交互 | GitHub、可选飞书、Web 管理面板 | 聊天、任务看板、设备连接和后台工作 |

上述比较用于理解定位，不能据此推断两者存在实现继承关系、功能等价或性能高低。

## 5. 开源与项目状态

原型采用 MIT 许可证。官方明确说明：

- 代码仅供研究参考，不是生产就绪的软件。
- 公开版本经过适配，与内部版本以及《Super-agent 设计探索：迈向更自主的 Agent 系统》描述的实现不完全相同。
- 文章中的部分设计是未来方向，尚未在该原型中实现。

评估具体能力时，应区分设计文章的目标、公开原型已实现的行为和实际部署结果。

| 材料或版本 | 含义 |
|---|---|
| 设计文章 | 描述整体架构与探索方向，包含尚未实现的设计 |
| 内部版本 | 团队内部使用的实现，与公开版本不完全相同 |
| 公开研究原型 | 经过发布适配、可运行的部分实现，以仓库代码和配置后的实际行为为准 |

## 6. 参考资料

- [官方仓库与 README](https://github.com/deepseek-ai/dsh-superagent-prototype)
- [设计理念](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/main/docs/000-philosophy.md)
- [领域模型](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/main/docs/001-domain.md)
- [实现架构](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/main/docs/007-architecture.md)
- [许可证](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/main/LICENSE)
