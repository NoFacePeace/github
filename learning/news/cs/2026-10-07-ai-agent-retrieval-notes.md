# AI 与 Agent 新闻及检索实践笔记

整理日期：2026 年 10 月 7 日。新闻讨论范围：北京时间 2026 年 10 月 6 日至 10 月 7 日截至聊天时。来源日期保留原站标注，只有日期的公告不作精确时区换算。

本文整理本次聊天中的新闻、概念解释和落地建议。厂商指标及论文结果保留其适用条件；实现示例与起步配置属于工程建议。特别注意：媒体报道、论文收录和产品首次发布是不同时间点。

## 1. AI 新闻概览

### 1.1 Mistral Large 4 公开预览

Mistral 于 10 月 6 日发布 Large 4 公开预览。官方称其为原生多模态模型，总参数约 1 万亿、激活参数约 490 亿，面向编程、Agent 和企业任务。API 可试用，权重计划月底开放。性能领先的说法应结合独立评测判断。

来源：[Mistral 官方公告](https://mistral.ai/news/mistral-large-4/)。

### 1.2 DeepSeek 大额融资报道

路透社于 10 月 6 日援引知情人士称，DeepSeek 本轮融资预计超过 800 亿元人民币，预计当月完成；报道援引 Bloomberg 信息称腾讯与宁德时代是主要出资方之一。融资尚未正式官宣，相关公司未回应置评请求。

来源：[路透社报道，由 Investing.com 转载](https://www.investing.com/news/economy-news/deepseek-set-to-net-over-12-billion-in-new-fundraising-source-says-4933887)。

### 1.3 Google 核电合作与英国医疗 AI 监管

Google 于 10 月 6 日宣布与 Constellation 合作，通过升级现有核电机组新增 890 MW 容量，预计在 2032 年底前落实，并使用 Gemini Enterprise 辅助优化电站运营。

英国政府同日接受全部 44 项医疗 AI 监管建议，强调全生命周期监测，并开放第三阶段 AI Airlock 监管沙盒申请；完整实施路线图预计于 2027 年春季公布。

来源：[Google 官方公告](https://blog.google/company-news/why-were-backing-americas-existing-nuclear-plants/)、[英国政府公告](https://www.gov.uk/government/news/government-backs-recommendations-of-nhs-doctors-led-ai-commission)。

## 2. Agent 协议与身份管理

### 2.1 Personal Agent Protocol

Meta、Sierra 于 10 月 6 日宣布与 Shopify、Stripe、Walmart、Genesys 等共同开发 Personal Agent Protocol，定义个人 Agent 如何代表用户与企业交互。

协议设计基于 OAuth：用户决定读写授权，企业决定允许的操作范围；Agent 可以通过网页、API 或企业自己的 Agent 完成任务。v0.1 规范计划本月稍后发布，不能把公告理解为完整规范已经可用。

来源：[Sierra 官方公告](https://sierra.ai/blog/introducing-personal-agent-protocol)。

### 2.2 Aembit 扩展个人 Agent 访问控制

Aembit 于 10 月 6 日宣布支持个人 Agent 的企业访问控制，为 Agent 提供独立身份，同时保留其代表的员工身份。能力包括运行时授权、短期凭证、集中审计，以及单独撤销 Agent 权限。公司称已向现有客户开放。

来源：[Aembit 公司新闻稿，由 StreetInsider 转载](https://www.streetinsider.com/Globe%2BNewswire/Aembit%2BExtends%2BAccess%2BControls%2Bto%2BPersonal%2BAI%2BAgents/27153487.html)。

## 3. Copilot Studio Hooks

### 3.1 概念与时间澄清

Hook 是在特定事件发生时执行回调或流程的既有机制。Git hooks、Webhooks 和框架生命周期回调都采用类似思路，Agent 工具中也早已有此机制。

本次消息涉及 Copilot Studio 中由 GitHub Copilot harness 驱动的 Agent 或工作流提供 Hooks 预览支持，不能泛化为所有 Copilot 产品新增 Hooks。

官方文档最后更新日期为 9 月 29 日，10 月 6 日是媒体报道日期。因此，无法据此确定首次上线时间，也不应将其描述成 10 月 6 日刚发明或刚推出的机制。

来源：[微软官方文档](https://learn.microsoft.com/en-gb/microsoft-copilot-studio/agents-experience/hooks-overview)、[10 月 6 日媒体报道](https://cloudwars.com/ai/need-ai-agents-to-run-workflows-with-no-exceptions-microsoft-has-a-hook-for-that/)。

### 3.2 工作方式与用途

Hook 绑定生命周期事件和工作流。事件发生时，平台调用工作流，传入事件详情，并把输出交回 Agent。

| 事件 | 用途 |
|---|---|
| 会话开始 | 加载用户、团队、业务背景 |
| 用户提交消息 | 标准化或替换输入 |
| 工具执行前 | 校验参数、修改参数、拒绝工具调用 |
| 工具成功后 | 调整结果、脱敏、记录审计 |
| 工具失败后 | 提供失败处理指导 |
| Agent 错误 | 指定重试、跳过或终止等错误处理 |

普通工具由模型决定是否调用；Hook 在对应事件发生时自动触发。退款 Agent 可以在退款工具执行前检查金额和审批记录，执行后写审计记录。

### 3.3 失败行为与工程边界

官方文档明确指出：Hook 工作流失败、超时或返回不可解析内容时，Agent 默认按没有返回结果继续执行。Pre tool use 是提供明确工具调用拒绝能力的事件。

因此，关键业务权限、金额上限等规则仍应在实际 API 或服务层强制执行。Hook 适合提前检查、上下文补充与审计，不能作为关键规则的唯一保障。

## 4. EmbeddingGemma 2

### 4.1 定位与工作原理

Google 于 10 月 6 日发布 EmbeddingGemma 2，采用 Apache 2.0 许可。它把文字、代码、图片、音频、视频及组合输入映射到统一向量空间，用于语义搜索、分类、聚类和 RAG 检索。

Embedding 是表示内容特征的数字向量。应用先为资料建立向量索引，再把查询转成向量，通过相似度找到相关内容。例如，使用文字查找相关代码实现、录音内容或视频片段。

它输出向量，不直接生成答案；回答和任务执行需要另外的生成模型或 Agent。

来源：[Google 官方公告](https://blog.google/innovation-and-ai/technology/developers-tools/embeddinggemma-2/)、[官方模型卡](https://ai.google.dev/gemma/docs/embeddinggemma/model_card_2)。

### 4.2 参数与本地部署

| 加载范围 | 参数量 |
|---|---:|
| 文本和代码 | 2.7 亿 |
| 文本和图像 | 4.4 亿 |
| 文本和音频 | 5.7 亿 |
| 完整多模态 | 7.4 亿 |

模型上下文窗口为 8,192 token。Google 报告，在 Pixel 11 Pro 的量化配置下，仅文本权重的活跃内存约 191 MB，完整模型约 567 MB。这不是整个应用的内存需求，索引、输入处理和运行时还会占用资源。

本地生成向量和检索可以离线完成。若后续使用云端模型回答，发送给云端的证据仍会离开本机。

### 4.3 相对上一代的变化

| 官方基准 | 上一代 | EmbeddingGemma 2 |
|---|---:|---:|
| 多语言文本 MTEB | 61.15 | 61.36 |
| 代码 MTEB | 68.76 | 78.68 |

主要变化是多模态能力和代码表现，普通多语言文本基准提升较小。上述数值是基准分数，不能直接解释为业务准确率。

### 4.4 接入要点

- 使用对应的任务前缀，区分搜索、代码检索、分类等用途。
- 默认输出 768 维，也支持截断为 512、256、128 维；截断后重新归一化，查询与资料使用相同维度。
- 128 维对多模态质量影响更明显，应先验证实际任务。
- 原始模型推理使用 `bfloat16` 或 `float32`。模型卡警告 `float16` 可能产生 NaN 或质量下降；量化部署需遵循相应运行时说明。
- 文件解析、切分、索引、权限过滤和来源定位由应用实现。

## 5. Agentic Retrieval 的含义与研究结果

### 5.1 什么是智能体检索

Agentic Retrieval 可以译为“智能体检索”：Agent 根据已有搜索结果决定下一步查什么，在证据足够、没有新线索或预算耗尽时停止。

普通检索通常是“问题 → 向量或关键词搜索 → 返回结果”。智能体检索增加反馈循环：“搜索 → 阅读结果 → 判断缺口 → 改写或拆分查询 → 再搜索”。

例如调查服务变慢，先查日志，发现数据库超时，再查慢查询和连接池情况。重点是根据新线索调整查询，而非重复执行同一次搜索。

### 5.2 NVIDIA 论文的方法

论文《Beyond Semantic Similarity: Performance and Costs of Agentic Retrieval for Complex Tasks》于 10 月 5 日提交 arXiv，10 月 6 日进入 Hugging Face Daily Papers。

其 NeMo Retriever Agent 使用 ReAct 循环和 `retrieve`、`think`、`final_results` 工具，先提供初始检索结果，再动态探索缺失文档。最终返回排序后的文档列表，实验评估的是检索质量。

来源：[论文全文](https://arxiv.org/html/2610.05750v1)、[Hugging Face 收录页面](https://huggingface.co/papers/2610.05750)。

### 5.3 收益与成本

使用相同 embedding 模型比较，论文报告：

| 基准 | 普通检索 nDCG@10 | Agent 检索，Opus 4.5 |
|---|---:|---:|
| ViDoRe v3：企业文档检索 | 64.36 | 69.22 |
| BRIGHT：推理密集检索 | 38.28 | 50.79 |

平均提升约 8.7 个点。nDCG@10 衡量前 10 个结果的相关性与排序质量，不代表回答准确率提升 8.7%。

ViDoRe v3 成本实验中，Agent 检索平均耗时约 107.4 秒，普通检索约 0.67 秒；每次查询累计消耗约 76.4 万输入 token、5,800 输出 token。这些数据属于特定配置，不能推广为所有 Agent RAG 的固定开销。

## 6. Agentic Retrieval 的最小落地方案

### 6.1 是否需要专用 Agent

需要负责检索决策的逻辑，但不必训练专用模型，也不必开发独立 Agent 产品。

- 已有 Agent：提供搜索工具和检索指导，由现有模型决定下一步查询。
- 已有普通 RAG：增加一个通用模型驱动的小循环，判断证据缺口并生成补充查询。
- 多个业务复用时：封装成 `retrieve(question) → evidence` 模块。

EmbeddingGemma 2 可作为底层向量检索组件，但不是必要依赖。关键词、向量或混合检索均可作为搜索工具。

### 6.2 搜索工具与指令

工具接口示例：

```python
search(query: str, top_k: int = 5) -> list[SearchHit]
```

每个命中结果至少包含稳定的片段 ID、标题、正文和来源；可增加更新时间。长文档可提供 `read_document(id)` 读取周边内容。权限过滤在检索服务中执行。

指令示例：

> 先搜索相关资料。若结果不足以回答问题，指出缺少的信息，并改写或拆分查询继续搜索。证据足够时结束；没有新线索时停止，并说明缺失部分。

这就是“提供 tool，结果不充分时调整查询再调用”的基本形式。轮数、耗时和 token 上限必须由程序约束。

### 6.3 执行循环

```text
用户问题
  → 初次检索
  → Agent 判断证据是否足够
      ├─ 足够：返回证据
      └─ 不足：生成补充查询
                → 检索、合并、去重
                → 再次判断
  → 达到上限或没有新线索：用已有证据结束
  → 生成带引用的答案
```

下面是实现结构示例，函数需由实际应用提供，并非完整可运行代码：

```python
evidence = search(question, top_k=5)

for _ in range(2):
    decision = planner(question, evidence)
    # finish + selected_ids，或 search + queries
    if decision["action"] == "finish":
        break

    queries = validate_queries(decision["queries"], limit=2)
    new_hits = parallel_search(queries, top_k=5)
    unseen_hits = exclude_existing(new_hits, evidence)
    if not unseen_hits:
        break
    evidence = merge_and_deduplicate(evidence, unseen_hits)

return answer_with_citations(question, evidence)
```

实际程序还需落实超时、token 预算、结构化输出校验、异常处理和来源验证。模型决策只在程序允许的范围内生效。

### 6.4 起步配置与评估

以下是工程建议，不是论文要求：

| 项目 | 建议初始值或策略 |
|---|---|
| 补充检索轮数 | 最多 2 轮 |
| 每轮查询数 | 最多 2 个，独立查询并行 |
| 每个查询结果数 | 5 条 |
| 保留上下文 | 最相关的 10～15 个片段，限制长度 |
| 停止条件 | 证据足够、无新增材料、预算耗尽 |
| 失败处理 | 返回已有材料，说明缺失信息 |

维护去重后的证据集合和已执行查询，避免反复传入全部历史材料。

挑选 30～50 个真实问题，包含简单查询、跨文档问题和无答案问题；用同一资料库比较普通检索、固定多查询检索和 Agent 检索。评估相关材料召回、引用支持程度、停止行为、延迟和任务总成本。

建议先作为“深入搜索”入口上线，验证收益后再自动路由复杂问题。检索完成后单独生成答案，便于区分检索问题与回答问题。

## 7. 三种机制的关系

| 机制 | 负责什么 |
|---|---|
| Hooks | 在执行生命周期节点运行固定流程，校验、补充上下文或记录结果 |
| EmbeddingGemma 2 | 把内容编码为向量，支持语义与跨模态检索 |
| Agentic Retrieval | 根据已有结果动态决定后续查询、证据选择和停止时机 |

三者可以组合：Agent 使用检索工具探索资料，底层使用 EmbeddingGemma 2 建索引，工具调用前后通过 Hook 做检查和记录。是否组合取决于业务需求，最小检索循环不依赖完整框架。
