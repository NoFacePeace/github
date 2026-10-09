# AI 与 Agent 新闻摘要

整理日期：2026 年 10 月 9 日。检索范围：北京时间 2026 年 10 月 8 日至 10 月 9 日截至本次整理。

本文汇总本次对话中核对过的 AI 与 Agent 新闻及相关技术说明，不代表全部新闻。来源日期保留原站标注；仅标日期的公告不作精确时区换算。产品能力和研究结果按官方材料理解，以下明确区分首次发布、版本更新与扩展上线。

## 1. Google 发布企业级 Gemini Agent

发布事件：美国时间 2026 年 10 月 8 日的 Gemini at Work 活动。官方页面本次读取标注 10 月 9 日，检索摘要标注 10 月 8 日，故保留这一日期差异。

Google 将 Gemini Agent 定位为统一的企业工作 Agent：处理问答、知识工作、内容生成和代码执行，并通过工具连接业务系统。公告强调云端持续执行、多 Agent 协作、记忆、模型选择，以及企业权限、审计和成本控制。

### 1.1 工作方式与集成

- 可以接收目标、定时任务或事件触发的工作，在云端持续执行。
- 可创建子 Agent 协作，并提供具有持续职责和独立身份的 coworker agent。
- 支持连接 Workspace、Slack、Git、Jira 等工具及 MCP 服务。
- 公告称可在 Gemini 与 Claude 模型间选择，未来计划支持更多模型。

工程关注点：如何在跨应用任务中保持上下文，并落实身份、授权、执行隔离和预算边界。公告不等于所有功能已经向所有账户开放。

### 1.2 是否开源

截至本次检索，未找到这次完整 Gemini Agent 产品的源码仓库或开源许可证；官方将其作为企业云产品介绍。Google 的 Agent Development Kit（ADK）是开源开发框架，可以用于自行构建 Agent，不能将 ADK 开源理解为完整 Gemini Agent 产品开源。

来源：[Google 官方公告](https://cloud.google.com/blog/products/ai-machine-learning/welcome-to-gemini-at-work-2026)、[ADK 官方介绍](https://developers.googleblog.com/en/agent-development-kit-easy-to-build-multi-agent-applications/)。

## 2. Codex 提高运行中引导的响应速度

官方公告日期：2026 年 10 月 8 日。

ChatGPT 桌面应用中的 Codex 正在推出 faster steering：任务运行时，用户补充信息、纠正方案或改变方向，Codex 能更快响应。设置中的 Follow-up behavior 可选择引导当前运行，或等待下一轮处理。

### 2.1 与原有引导的区别

公告说明的是响应更快，没有公布加速幅度、更新前后的处理流程或具体实现。因此不能确定它改动了消息队列、模型推理还是工具调度，也不能理解为新消息立即停止当前操作。

### 2.2 相关 API 机制与工具执行边界

官方 Mid-turn steering API 文档提供更具体的机制：通过 WebSocket 的 `response.steer` 发送补充指令；输入被接收后排队，服务端完成当前输出项及已经运行的托管工具工作，再自动带着新指令继续生成。

该机制不会撤销先前动作，也不会取消已经启动的工具。后续调用可在继续生成时根据新指令重新决定，但文档没有保证所有已排队、未启动的调用立即暂停。

这份 API 机制说明与 10 月 8 日桌面更新没有被官方明确关联，不能据此断言桌面更新采用了同一实现。

来源：[ChatGPT 官方更新日志](https://help.openai.com/en/articles/6825453-chatgpt-release-notes)、[Mid-turn steering 技术文档](https://developers.openai.com/api/docs/guides/steering)。

## 3. MIMESIS：用更像真人的模拟用户训练 Agent

论文首次提交日期：2026 年 10 月 7 日；本次阅读版本于 10 月 8 日更新。这是一项研究更新。

普通助手模型扮演用户时可能过于配合，从而低估交互难度。MIMESIS 学习真实用户行为，作为 Agent 的训练与评估环境。

### 3.1 训练流程

模拟用户基于 Qwen3.5 的 4B、9B 模型：先学习真实对话中的下一条用户回复，再用 ThoughtTrace 的用户自述动机监督私有思考，最后通过 GRPO 强化学习 13 类行为。行为奖励关注表现、时机、自然度及任务一致性。

随后冻结模拟用户，与 Qwen3-8B Agent 多轮交互，用任务奖励训练 Agent。CSD 方法进一步将模拟用户的思考和反应转换成教练反馈，提供 token 级监督；实际运行的 Agent 只能看到公开对话。

### 3.2 评估与结果

模拟用户评估关注行为保真、对话节奏和真人交互难度校准。Agent 在 8 个环境、9 个未参与训练的用户模型上测试，其中 3 个环境也未参与训练。

| Agent 训练方式 | 平均任务分数 |
|---|---:|
| GPT-5.5 模拟用户 + GRPO | 26.10 |
| MIMESIS-9B + GRPO | 29.54 |
| MIMESIS-9B + GRPO + CSD | 31.09 |

结果支持对未见模拟用户的泛化改善；这些分数不能直接视为真实用户成功率，也没有证明上线后真人体验同等改善。

来源：[论文摘要与版本记录](https://arxiv.org/abs/2610.09484)、[论文全文](https://arxiv.org/html/2610.09484v2)。

## 4. ChatGPT 的 GPT-6 与交互式 UI 扩展上线

首次公告日期：2026 年 10 月 7 日。官方安排从 10 月 8 日起扩展到 Free 和 Go，因此本期记录的是扩展上线。

Intelligent UI 可以将文字、图表、比较视图和交互组件结合在回答中，例如计算器、费用分摊工具和小游戏。模型还可以先展示部分答案，在继续思考或使用工具后补充结果。

官方明确说明，Work 和 Codex 的底层模型不随这次 ChatGPT Chat 发布改变。

### 4.1 图片、SVG 与 token 用量

- **交互组件与图表：** 不一定调用图片生成模型。官方没有公布 Intelligent UI 相比纯文本的额外 token 用量或专门计费规则。
- **生成位图：** API 的图片生成有图片 token 和费用；通过 Responses API 调用时还包括主模型用量。ChatGPT 订阅额度不能直接按 API 价格估算。
- **SVG：** 直接输出 SVG 源码通常属于文本输出，消耗文本 token；浏览器渲染 SVG 本身不等于调用图片生成模型。若转换为位图后提交给视觉模型，则按相应图片输入规则处理。

因此，回答中出现图形不代表必然消耗图片生成额度。具体取决于内容生成与传入模型的方式；本次发布未明确 Intelligent UI 自动生成图片的额度归属。

来源：[ChatGPT 官方更新日志](https://help.openai.com/en/articles/6825453-chatgpt-release-notes)、[API 图片生成说明](https://developers.openai.com/api/docs/guides/image-generation)、[图片输入与视觉说明](https://developers.openai.com/api/docs/guides/images-vision)、[订阅用量说明](https://learn.chatgpt.com/docs/pricing)。

## 5. 阅读优先级

| 新闻 | 建议关注的问题 |
|---|---|
| Gemini Agent | 跨应用任务、长期记忆、多 Agent 协作与企业治理如何组合 |
| MIMESIS | 模拟用户是否扭曲任务难度，训练效果能否迁移到不同用户 |
| Codex steering | 新指令何时生效，已运行工具与后续动作的边界是什么 |
| Intelligent UI | 哪些问题适合交互表达，以及不同生成方式如何影响用量 |

结合 Agent 工程学习方向，优先阅读 Gemini Agent 官方公告和 MIMESIS 论文，分别研究企业产品架构与交互训练评估。
