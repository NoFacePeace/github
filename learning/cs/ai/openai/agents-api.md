# Agents API

## 1. Computer use

### 1.1 概念与职责

Computer use 让 Agent 通过浏览器界面浏览网站、收集信息和测试应用。本文介绍 Agents API 的托管浏览器模式。

模型负责观察界面并决定下一步；OpenAI 服务端提供浏览器环境并执行操作。应用负责启动会话、处理事件和用户交互，并验证结果。它属于 Agent 运行服务的能力，包含模型推理和配套执行系统。

### 1.2 与 Responses API 的区别

| 对比项 | Responses API computer use | Agents API 托管浏览器 |
| --- | --- | --- |
| 浏览器环境 | 应用提供并控制 | OpenAI 托管 |
| 操作执行 | 应用执行 `computer_call` 中的动作 | 服务端执行浏览器操作 |
| 状态反馈 | 应用回传截图作为 `computer_call_output` | 托管运行系统处理观察与操作循环 |
| 应用职责 | 实现操作与截图循环 | 管理会话、处理审批和登录、核验结果 |

两者都使用模型判断操作，但执行系统的归属不同。仅部署一个模型或兼容请求格式，不会自动获得托管浏览器。

### 1.3 工作流程

1. 创建浏览器会话并保存会话 ID。
2. 提交任务并监听事件。
3. 处理网站访问审批和需要的登录交互。
4. 等待 Agent 完成，检查结果和浏览器活动。
5. 使用结束后删除会话；断线时先恢复原会话，避免重复执行任务。

创建会话本身不等于开始执行任务。

### 1.4 配置示例

以下示例创建浏览器会话，不包含后续任务提交和事件处理。Beta 接口接入时应核对当前文档。

```http
POST https://api.openai.com/v1/agents/sessions
Authorization: Bearer <OPENAI_API_KEY>
Content-Type: application/json
OpenAI-Beta: agents=v1
```

```json
{
  "agent": {
    "model": "gpt-6-astra",
    "instructions": "Read public documentation and report the page title and URL. Do not sign in or modify website data.",
    "tools": [
      {
        "type": "computer_use",
        "include_screenshots": true
      }
    ]
  },
  "environment": {
    "type": "openai_hosted",
    "desktop": { "enabled": true },
    "network": { "access": "enabled" }
  }
}
```

### 1.5 网站访问审批与登录

启用网络不代表批准所有网站。访问新的网站源时需要用户批准；应用监听 `agent.session.requires_action`，读取会话的 `required_actions`，展示请求并提交用户决定。

登录通过专用交互处理，用户在聊天之外输入凭据，应用按请求字段提交。不要把密码等凭据放入普通消息、工具结果或日志。

网站访问许可不等于逐项操作确认。若业务要求强制确认购买、删除等动作，需要在可执行资源或应用控制的运行环境中落实限制。

### 1.6 使用场景与参考资料

可用于网页信息采集、界面测试和通过 UI 操作业务系统。选择托管模式还是自行执行，取决于是否需要直接控制浏览器环境和每一步操作；这是架构选择建议。

- [OpenAI Docs：Agents API Computer use](https://developers.openai.com/api/docs/guides/agents-api/tools/computer-use)
- [OpenAI Docs：Responses API Computer use](https://developers.openai.com/api/docs/guides/tools-computer-use)
