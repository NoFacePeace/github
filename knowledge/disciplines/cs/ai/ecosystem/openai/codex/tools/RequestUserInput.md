# Codex 向用户提问

Codex 的内置提问工具让模型提交结构化问题，由运行时将问题交给客户端显示。本文依据开源 Codex Core、App Server 和终端客户端整理，核对日期为 2026-11-26。

## 1. `request_user_input` 等待答案

1. 模型提交 `questions`。每道题有用于匹配答案的 `id`、题干和选项等字段，工具处理器校验参数后发起提问。
2. Core 在当前轮次的 `pending_user_input` 中保存 Rust `oneshot` 通道的发送端，工具调用等待接收端。App Server 向客户端发送 `item/tool/requestUserInput` JSON-RPC 请求，携带 `threadId`、`turnId`、`itemId` 和问题列表。客户端据此显示提问界面。
3. 用户提交后，客户端用原 JSON-RPC 请求的 `id` 回包。以下示例是回包中的 `result`，其中 `storage` 是问题 ID。

   ```json
   {"answers":{"storage":{"answers":["SQLite"]}}}
   ```

4. Core 将答案发给等待中的工具处理器，处理器把答案序列化为工具结果，模型据此继续当前轮次。App Server 发出 `serverRequest/resolved`，客户端可据此关闭提问界面。

## 2. `request_user_input_async` 发出问题后继续

异步工具接受一道或多道题，每道题包含 `title` 和可选的 `options`。Core 把问题附在异步 Agent 消息上，发出消息事件后立即返回 `{"accepted":true}`。这个结果表示提问已发出。此时工具调用已经完成，提问仍待用户回复。

用户稍后提交的答案作为后续用户输入进入会话。Agent 可以先处理与答案无关的工作，等收到输入后再执行依赖答案的步骤。

## 3. 工具可用性

工具是否提供给模型由协作模式、功能开关和模型目录等因素决定。公开实现中，`request_user_input` 在 Plan 模式下会将请求标记为阻塞，部分配置也可以在 Default 模式暴露该工具。`request_user_input_async` 的注册会检查模型目录声明的能力。实际可用的工具以当前会话为准。

## 4. 参考资料

- [阻塞工具处理器](https://github.com/openai/codex/blob/main/codex-rs/core/src/tools/handlers/request_user_input.rs)
- [轮次状态中的待回答请求](https://github.com/openai/codex/blob/main/codex-rs/core/src/state/turn.rs)
- [App Server 协议](https://learn.chatgpt.com/docs/app-server#toolrequestuserinput)与[回答回传测试](https://github.com/openai/codex/blob/main/codex-rs/app-server/tests/suite/v2/request_user_input.rs)
- [终端提问界面](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/request_user_input/mod.rs)
- [异步工具处理器](https://github.com/openai/codex/blob/main/codex-rs/core/src/tools/handlers/request_user_input_async.rs)与[工具注册](https://github.com/openai/codex/blob/main/codex-rs/core/src/tools/spec_plan.rs)
