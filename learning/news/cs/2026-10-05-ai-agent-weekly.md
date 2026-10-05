# AI 与 Agent 新闻周报

覆盖时间：2026 年 9 月 29 日至 10 月 5 日。整理日期：2026 年 10 月 5 日。下文日期采用来源标注的日期。

本周关注新模型、持续运行的 Agent、多 Agent 协作、事件触发和企业部署。主要变化是产品逐步支持持续执行任务，同时权限控制和监管受到更多关注。模型性能表述来自厂商公告，不能直接视为独立测试结论。

## 1. 本周 AI 新闻

### 1.1 OpenAI 发布 GPT-6.1 Sol

9 月 29 日，OpenAI 发布 GPT-6.1 Sol，面向复杂编程、电脑操作和专业工作。官方称其以较低成本提供接近 GPT-6 Astra 的性能。

标准 API 价格在输入不超过 272K token 时，按每百万 token 计费：输入 2 美元、缓存输入 0.10 美元、输出 10 美元。该模型同时支持 Responses API 中的 Multi-agent beta。

对开发者的启示：可以用真实任务比较完成质量、耗时和总成本，评估是否适合日常编程与自动化工作。

来源：[OpenAI API 更新记录](https://developers.openai.com/api/docs/changelog)、[GPT-6.1 Sol 模型文档](https://developers.openai.com/api/docs/models/gpt-6.1-sol)。

### 1.2 OpenAI 发布 Dots 与新的工作功能

9 月 29 日，OpenAI 宣布 Dots。它是持续运行的 Agent，拥有自己的电脑，可在连接的应用中工作，在对话之间继续推进任务，并把结果交给用户审阅。需要用户判断时，它会主动联系用户。目前逐步向符合条件的账户开放。

同场公布的 ChatGPT Space 支持将文件和 Pages 放在共同工作空间中。Codex 的可复用云端环境则允许预先准备开发环境，使新任务在隔离的工作空间中启动，并支持从网页、手机或桌面继续云端任务。

来源：[ChatGPT Learn DevDay 2026 发布说明](https://learn.chatgpt.com/docs/whats-new/devday-2026)。

### 1.3 Google 发布 Gemini 4 Argon

9 月 30 日，Google 发布 Gemini 4 Argon，定位于复杂、长流程的软件工程、企业知识工作和网络安全防御。官方将输出 token 上限从此前的 64K 提升到 1M，并表示模型能够自主发现、验证和修补软件漏洞。

初期通过 Fairwind 项目向可信网络防御团队开放。Google 表示将继续测试和加强防护，再扩大到开发者、企业和消费者。

对开发者的启示：长流程编程任务和安全修复值得关注，但当前开放范围有限，公告中的性能仍需结合实际使用验证。

来源：[Google 官方公告](https://blog.google/innovation-and-ai/models-and-research/gemini-models/gemini-4-argon/)。

### 1.4 美国 FTC 调查 AI 智能体风险

9 月 30 日，路透社报道，美国联邦贸易委员会（FTC）正在对 Anthropic、OpenAI 等 AI 实验室展开行业调查，关注其技术对消费者的潜在风险，尤其是 Agent 越权和失控行为。报道援引 FTC 官员称，机构计划要求提供资料并要求相关高管作证。

这是一项调查进展，不能视为已经认定相关公司违法。

来源：[路透社报道，由 MarketScreener 转载](https://ca.marketscreener.com/news/ftc-opens-probe-into-ai-giants-including-anthropic-and-openai-new-york-post-reports-ce785ad2d08bf22d)。

### 1.5 Anthropic 投入 1 亿美元培养企业 AI 工程师

10 月 2 日，Anthropic 推出 Claude Frontier Academy，承诺投入 1 亿美元，计划到 2027 年底培养 1 万名 Frontier Deployed Engineers。首批参与机构包括埃森哲、贝恩、德勤、麦肯锡和摩根士丹利等。

培训包含企业部署模拟、实践考核，以及在自身组织中推进真实 Claude 项目的 12 周实践。项目由组织提名参与。

对开发者的启示：企业 AI 落地需要把模型接入业务、完成安全审查并持续交付的工程能力。

来源：[Anthropic 官方公告](https://www.anthropic.com/news/claude-frontier-academy)。

## 2. Agent 开发重点

### 2.1 多 Agent 协作

GPT-6.1 Sol 在 Responses API 中支持 Multi-agent beta，允许模型将工作委派给子 Agent。

应用判断：研究、编程和资料整理中可独立拆分的任务，适合尝试这种机制。实际收益需要评估任务拆分质量、上下文传递、结果整合和额外 token 成本。

来源：[OpenAI API 更新记录，9 月 29 日条目](https://developers.openai.com/api/docs/changelog)。

### 2.2 托管浏览器执行

9 月 29 日，OpenAI Agents API 新增 computer use。Agent 可在 OpenAI 托管的浏览器中完成任务，网站访问审批和登录由接入应用处理。

应用判断：这为网页操作类 Agent 提供了托管执行环境。产品仍需清楚定义访问范围和涉及用户判断的操作。

来源：[OpenAI API 更新记录，9 月 29 日条目](https://developers.openai.com/api/docs/changelog)。

### 2.3 MCP Events 与事件触发

DevDay 公告介绍了 MCP Events：用户可让 ChatGPT 监听 MCP 服务的更新，在匹配事件到达时采取行动。该能力要求 MCP 2.0，并支持草案 MCP Events 规范中的 webhook 交付机制。

应用判断：可以探索收到工单、代码更新或业务状态变化后触发处理的工作流。具体可用范围和实现方式应以接入文档为准。

来源：[ChatGPT Learn DevDay 2026 发布说明](https://learn.chatgpt.com/docs/whats-new/devday-2026)。

### 2.4 编程与安全 Agent

Codex 新增可复用云端环境；Security Cloud 插件进入研究预览，支持扫描连接的 GitHub 仓库或监控新提交，并在创建草稿 PR 前审阅发现、验证证据与补丁。

应用判断：编程 Agent 正在覆盖从开发环境准备到安全检查和修复提案的更多环节。实际落地应重视可复现环境、变更审阅和验证证据。

来源：[ChatGPT Learn DevDay 2026 发布说明](https://learn.chatgpt.com/docs/whats-new/devday-2026)。

## 3. 趋势判断与关注顺序

以下为基于上述新闻的判断：

- **持续工作成为产品方向。** Dots 和云端开发环境使任务可以跨对话、跨设备继续推进。
- **Agent 基础设施继续完善。** 多 Agent 委派、托管浏览器和事件触发，分别补充任务拆分、实际执行和自动启动能力。
- **权限与审计成为落地关键。** 执行时间越长、可访问系统越多，越需要控制范围、记录行动并支持人工介入。
- **企业部署需要工程与业务能力。** Anthropic 的培训计划将真实项目交付放在核心位置。

如果以开发 Agent 为目标，建议先关注 Multi-agent API、托管 computer use 和 MCP Events，再研究持续运行任务的状态管理、权限控制和失败恢复。以上属于学习与实现建议，并非厂商已验证的通用最佳实践。
