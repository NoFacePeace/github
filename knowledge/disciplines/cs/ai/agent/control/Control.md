# Agent 任务控制

整理任务如何启动、如何设定目标、如何在执行过程中调整，以及如何拆分和协调工作。

## 1. 主题

- [触发机制](trigger/Trigger.md)
- [交互机制](interaction/Interaction.md)
- [编排](orchestration/Orchestration.md)
- [目标](goal/Goal.md)

## 2. 概念区分

理解 Agent 的工作方式，可以先区分以下问题，而不把所有概念统称为互斥的“模式”。不同产品的命名和实现可能不同，也没有统一固定的分类数量。

| 问题 | 对应概念 | 内容归属 |
|---|---|---|
| 让 Agent 做什么？ | Ask、Plan、Execute，部分产品另有 Edit | 交互机制中的工作模式 |
| Agent 怎么组织步骤？ | Workflow、ReAct、Plan-and-Execute，以及可叠加的 Reflection | 编排 |
| 谁来做？ | 单 Agent、多 Agent 分工 | 编排 |
| 要达成什么、什么时候结束？ | Goal、完成条件、持续推进 | 目标 |
| 哪些行动需要确认？ | 授权与审批 | 交互机制中的审批 |

例如，一个 Agent 可以在 Execute 工作模式下，先制定计划，再用 ReAct 完成各步骤，并在完成后反思和修正。工作模式、执行方式、协作结构和自主权限需要分别理解。
