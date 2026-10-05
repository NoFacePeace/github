# Models API

## 1. Multi-agent

### 1.1 概念与接口归属

Multi-agent 允许主 Agent 拆分任务，创建子 Agent 并汇总结果。本文讨论的是 Responses API（`POST /v1/responses`）中的托管多 Agent 功能，与独立的 Agents API 区分。

它属于模型调用接口提供的服务端编排能力：模型决定如何分工，OpenAI 的运行程序负责调度。模型本身不会执行程序，也不会直接创建运行进程。

### 1.2 工作流程

以代码审查为例，主 Agent 可以让子 Agent 分别检查正确性、安全性和测试覆盖，最后合并结论。

```text
应用提交任务
    ↓
主模型决定拆分与委派
    ↓
OpenAI 服务端创建子 Agent 上下文并调用模型
    ↓
子 Agent 分析并返回结果
    ↓
主模型汇总回答
```

子 Agent 可以使用同一个模型，各自维护独立上下文。“多个 Agent”不等于“多个不同模型”。

### 1.3 返回结果与工具执行

应用不需要收到 Agent 名单后再逐个调用。服务端处理创建、通信和等待等协作动作；返回中可能包含过程记录和最终回答。

| 返回项 | 应用如何处理 |
| --- | --- |
| `multi_agent_call` | 托管协作动作，无需应用执行或提交结果 |
| `function_call` | 应用执行自定义函数，再提交对应的 `function_call_output` |
| 主 Agent 的最终消息 | 提取并展示给用户 |

例如，分析已经提供的文档可以直接完成；如果子 Agent 需要调用应用定义的数据库查询函数，应用仍需执行查询并回传结果。因此，一次完整任务可能跨越多次 API 请求。

### 1.4 与 Tool 的关系

从模型决策角度看，委派类似使用工具，但交出去的是一个可包含多步推理和工具调用的子任务。

- 普通工具：执行明确操作，例如读取文件。
- 子 Agent：完成目标，例如检查权限逻辑，并自行决定分析步骤。

托管委派动作由 OpenAI 服务端执行；应用定义的业务工具由应用执行。

### 1.5 开启方式

以下为请求示例。Beta 接口可能变化，接入时应核对当前官方文档和模型支持范围。

```http
POST https://api.openai.com/v1/responses
Authorization: Bearer <OPENAI_API_KEY>
Content-Type: application/json
OpenAI-Beta: responses_multi_agent=v1
```

```json
{
  "model": "gpt-6.1-sol",
  "input": "Review this diff with subagents for correctness, security, and tests. Reconcile their findings. <diff>...</diff>",
  "multi_agent": {
    "enabled": true,
    "max_concurrent_subagents": 3
  }
}
```

开启后模型可以委派，是否使用可通过指令约束。并发限制覆盖整个子 Agent 树，不包含主 Agent。

### 1.6 兼容接口与自部署模型

支持 OpenAI 请求格式，不代表实现了托管 Multi-agent。其他服务要兼容此功能，还需实现对应的调度和返回行为，应以其文档为准。

自部署模型通常只提供推理服务。开发者可以搭配 Agent 框架或自己编写调度程序，实现任务拆分、独立上下文、工具执行与结果回传。这是自行实现的多 Agent 系统，不能自动等同于兼容 OpenAI 的托管接口。

### 1.7 使用取舍与参考资料

可独立拆分的研究、审查和探索任务适合尝试多 Agent。步骤强依赖、任务很小或频繁修改同一资源时，应先考虑单 Agent。增加子 Agent 可能提高 token 用量，是否更快、更准确需要实际评估。

官方依据：[OpenAI Docs：Responses API Multi-agent](https://developers.openai.com/api/docs/guides/responses-multi-agent)。本文中的兼容性分析与自部署方案属于架构解释。
