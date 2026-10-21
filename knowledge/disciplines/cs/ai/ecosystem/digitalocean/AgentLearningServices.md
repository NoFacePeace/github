# DigitalOcean Agent Learning Services

Agent Learning Services 是 DigitalOcean 提供的一组 Agent [可观测性](../../agent/quality/observability/Observability.md)与[评估](../../agent/quality/evaluation/Evaluation.md)服务，包含 Insights、Signals 和 Evaluations & Simulations。它帮助开发者记录执行过程、发现问题，并在上线前测试修改是否有效。

## 1. 核心组成

| 组成 | 作用 | 所属方面 |
|---|---|---|
| Insights | 记录和查看 Agent 的执行过程 | 可观测性 |
| Signals | 从实际运行会话中筛选潜在问题，例如用户放弃、工具反复失败、超时 | 可观测性与问题分析 |
| Evaluations & Simulations | 按测试场景评估表现，并模拟多轮用户对话验证修改 | 评估与测试 |

## 2. 如何理解“学习”

“学习”主要指利用运行证据持续改进 Agent，不意味着每次对话都会自动更新模型参数。一个常见流程是：

```text
Agent 实际运行 → 记录执行过程 → 发现问题会话
                                   ↓
上线验证后的修改 ← 模拟测试与评估 ← 修改提示词、工具或流程
```

例如，客服 Agent 连续调用退货工具失败，用户最终放弃。Signals 标记这段会话；开发者修正工具参数或提示词后，Simulations 用模型扮演用户，再次模拟退货对话，检查问题是否解决。

这不要求另一个 Agent 实时盯着或接管回答。Signals 主要分析运行记录；Simulations 则在测试环境中模拟交互。运行数据也可用于后续微调或强化学习，但模型训练是单独的改进步骤。

## 3. 使用流程

### 3.1 启用 Signals

1. 在 DigitalOcean Managed Agents 上运行 Agent。
2. 在控制台的 Data & Learning 中打开 Signals，也可从 Harness Runtime 的 Signals 入口进入。
3. 为目标 Agent 启用 Signals；官方说明数据处理可能需要最多一小时才开始。
4. 查看用户交互、工具执行和环境问题的趋势，打开被标记的会话。
5. 将有代表性的会话导出为 Evaluations 测试场景。

### 3.2 运行 Simulations

1. 在控制台打开 Evaluations。
2. 在 Simulation Scenarios 中选择现有场景、用自然语言描述新场景，或从 Signals 导入。
3. 点击 Run Simulation，选择待测 Agent 或模型、用户模拟模型及评分指标。
4. 设置每个场景的运行次数和对话轮数上限。
5. 查看评分和完整执行轨迹，将重要场景保留为后续回归测试。

## 4. 与 Data Services 的关系及开放状态

[Agent Data Services](AgentDataServices.md) 提供资料、会话和执行记录的数据基础；Learning Services 使用运行记录发现问题并验证改进。

截至 2026 年 10 月 19 日，Insights、Signals、Evaluations & Simulations 处于公开预览；Data Services 和 Inference Router optimizer 处于私有预览。官方使用流程主要围绕 DigitalOcean 的托管 Agent 和推理基础设施展开。

官方发布说明中，当前自动优化能力包括根据评估结果选择模型；将同样的自动改进循环扩展到 Agent 工作流仍属于后续计划，不能理解为完整流程已经可以自动修改和上线。

## 5. 参考资料

- [Agent Learning 与 Data Services 发布公告](https://investors.digitalocean.com/news/news-details/2026/DigitalOcean-Launches-Agent-Learning-and-Data-Services-for-Recursively-Self-Improving-AI-Agents/default.aspx)
- [Signals 与 Simulations 介绍及操作步骤](https://www.digitalocean.com/blog/introducing-signals-and-simulations)
