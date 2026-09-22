# Agent 评测

## 1. 评测方法

评测不应只看最终答案，还要观察 Agent 的执行过程：

| 维度 | 关注点 |
| --- | --- |
| 任务成功率 | 是否满足任务目标和验收条件 |
| 结果正确性 | 事实、计算、代码和引用是否正确 |
| 工具使用 | 是否选择正确工具，参数和调用顺序是否合理 |
| 效率 | 步数、延迟、Token 消耗和工具成本 |
| 稳定性 | 相同输入下结果是否可接受，失败后能否恢复 |
| 安全性 | 是否越权、泄露信息或执行未授权副作用 |

应使用代表性任务集、失败案例和回归测试持续评估，并保留完整执行轨迹。

## 2. 评测分层

Agent 评测通常分为三层：

| 层次 | 关注点 | 常见实现 |
| --- | --- | --- |
| 结果评测 | 任务是否真正完成，最终状态是否正确 | 单元测试、数据库状态检查、文件校验、业务规则 |
| 轨迹评测 | Agent 是否选择了正确的工具和执行路径 | 工具调用检查、参数校验、步骤顺序、重试和循环检测 |
| 运行评测 | 在不同版本和真实流量下是否稳定 | 数据集回归、Trace 分析、线上采样、人工标注 |

确定性规则应优先用于可自动验证的结果，例如代码测试、JSON Schema、文件内容和数据库状态；LLM Judge 适合补充评估语义质量、完整性和主观标准，但关键场景仍应使用人工抽检校准。

## 3. 工具定位

不同工具解决的问题不同，不能简单按“评测框架”归为一类：

| 类型 | 主要职责 | 代表工具 |
| --- | --- | --- |
| Benchmark Harness | 准备沙箱环境，运行 Agent 和标准任务，执行验证器 | Harbor |
| 评测框架 | 编写数据集、指标、断言和回归测试 | DeepEval、Promptfoo、OpenAI Evals、Inspect AI |
| Trace 与评测平台 | 记录运行轨迹，管理数据集，进行实验对比和线上监控 | LangSmith、Braintrust |

## 4. 主流工具

### 4.1 Harbor

[Harbor](https://github.com/harbor-framework/harbor) 是由 Terminal-Bench 团队维护的 Agent 评测和优化框架，重点是统一运行环境、任务数据集、Agent 适配器和结果验证器。它可以在本地 Docker 或云环境中运行 Claude Code、OpenHands、Codex CLI、Aider 等 Agent，并执行 Terminal-Bench、SWE-bench 等任务集。

**定位**：Benchmark Harness 和沙箱任务执行器。

**适用场景**：

- Coding Agent、Terminal Agent 和 Browser Agent。
- 需要文件系统、代码仓库或终端的长流程任务。
- 需要在相同环境下比较多个 Agent、模型或 Harness。
- 需要隔离副作用并使用测试脚本验证最终状态。

**官方资源**：

- [Harbor GitHub](https://github.com/harbor-framework/harbor)
- [Harbor 文档](https://www.harborframework.com/docs)

### 4.2 DeepEval

[DeepEval](https://deepeval.com/) 是 Python 评测框架，支持通过 pytest 编写 Agent 评测，并对 Agent 的最终结果、工具调用、子 Agent 和其他组件进行评估。它适合把评测接入本地开发和 CI/CD。

**定位**：代码优先的 Agent 和 LLM 评测框架。

**适用场景**：

- 使用 Python 和 pytest 的项目。
- 需要自定义指标、LLM Judge 和组件级评测。
- 需要将 Agent 评测作为回归测试运行在 CI 中。

**官方资源**：

- [DeepEval 官网](https://deepeval.com/)
- [Agent 评测文档](https://deepeval.com/docs/getting-started-agents)

### 4.3 Promptfoo

[Promptfoo](https://www.promptfoo.dev/) 是配置化的 LLM 和 Agent 评测工具，通常使用 YAML 定义模型、Prompt、测试用例和断言，再通过 CLI 执行评测。它适合快速比较不同模型、Prompt、Provider 和 Agent 版本。

**定位**：配置化评测、模型对比和 CI 工具。

**适用场景**：

- 需要快速建立多模型或多 Prompt 对比。
- 希望使用 YAML 和 CLI 管理评测配置。
- 需要将断言、红队测试和回归测试接入 CI。

**官方资源**：

- [Promptfoo 官网](https://www.promptfoo.dev/)
- [Getting Started](https://www.promptfoo.dev/docs/getting-started/)

### 4.4 OpenAI Evals

[OpenAI Evals](https://github.com/openai/evals) 是用于评估 LLM 和 LLM 系统的开源框架，同时提供可复用的 Benchmark Registry。它支持使用公开数据集，也支持为自己的业务数据编写自定义评测和模型评分器。

**定位**：通用 LLM 系统评测框架和 Benchmark Registry。

**适用场景**：

- 需要维护标准化数据集和自定义 Eval。
- 需要比较模型版本或 Prompt 版本。
- Agent 主要通过模型调用、Completion Function 或工具调用完成任务。

**官方资源**：

- [OpenAI Evals GitHub](https://github.com/openai/evals)
- [OpenAI Evals](https://evals.openai.com/)

### 4.5 Inspect AI

[Inspect AI](https://inspect.aisi.org.uk/) 是面向 AI 系统评测的开源框架，使用 Dataset、Solver 和 Scorer 组织一次评测，并支持 Agent、工具调用和沙箱环境。它适合研究型评测、代码任务、安全测试和需要自定义执行器的场景。

**定位**：可扩展的研究和安全评测框架。

**适用场景**：

- Coding Agent、工具型 Agent 和多步任务。
- 需要沙箱、定制 Solver 或定制 Scorer。
- 需要进行安全、能力和行为边界评测。

**官方资源**：

- [Inspect AI 文档](https://inspect.aisi.org.uk/)
- [Inspect AI GitHub](https://github.com/UKGovernmentBEIS/inspect_ai)

### 4.6 LangSmith

[LangSmith](https://www.langchain.com/langsmith/evaluation) 是 LangChain 生态的 Agent 和 LLM 评测平台，支持数据集、离线评测、Trace、人工标注、启发式检查、LLM Judge、成对比较和线上评测。

**定位**：Trace、实验管理和生产评测平台。

**适用场景**：

- 使用 LangChain 或 LangGraph 的 Agent。
- 需要查看完整调用链、工具调用和失败原因。
- 需要将离线数据集评测和线上生产 Trace 连接起来。

**官方资源**：

- [LangSmith 评测](https://www.langchain.com/langsmith/evaluation)
- [LangSmith 评测文档](https://docs.langchain.com/langsmith/evaluation)

### 4.7 Braintrust

[Braintrust](https://www.braintrust.dev/) 是面向 AI 应用和 Agent 的托管评测平台，支持在代码或界面中定义任务和测试用例，并将 Trace、数据集、评分器和实验结果关联起来。

**定位**：团队协作、实验管理和生产评测平台。

**适用场景**：

- 需要多人协作管理数据集和评测结果。
- 需要比较多个 Agent 或模型版本。
- 需要从生产 Trace 中发现失败案例并转为新的回归测试。

**官方资源**：

- [Braintrust 官网](https://www.braintrust.dev/)
- [Agent 评测文档](https://www.braintrust.dev/learn/ai-agent-evaluation/v0)

## 5. 选型建议

可以按 Agent 类型和工程目标选择：

| 需求 | 优先考虑 |
| --- | --- |
| 普通业务 Agent，使用 Python 和 pytest | DeepEval |
| 多模型、Prompt 和 Provider 对比 | Promptfoo |
| LangChain 或 LangGraph 项目 | LangSmith |
| 代码、终端和浏览器 Agent | Harbor、Inspect AI |
| 研究型、安全型或需要自定义沙箱 | Inspect AI |
| OpenAI 模型和自定义 Eval | OpenAI Evals |
| 生产 Trace、团队协作和线上评测 | LangSmith、Braintrust |

工具可以组合使用。例如，使用 Harbor 在隔离环境中运行 Coding Agent，用测试脚本判断任务是否完成，再将执行 Trace 发送到 LangSmith 或 Braintrust 分析；DeepEval 或 Promptfoo 则可以补充业务指标和 CI 回归测试。
