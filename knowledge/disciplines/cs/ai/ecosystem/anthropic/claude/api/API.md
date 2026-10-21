# Claude API

Claude API 是 Anthropic 提供的模型调用接口。Anthropic 是公司名称，Claude 是模型与产品品牌；开发者通过 API 调用具体的 Claude 模型。

## 1. 接口

- [Messages API](Messages.md)：通过消息调用模型，支持文本生成、工具调用与流式输出。

## 2. 与 Agent 的关系

模型 API 接收上下文并返回模型输出。使用客户端工具时，开发者负责执行工具、回传结果并继续调用模型，从而构成 Agent Loop。

模型接口、工具执行与界面展示可以分层实现。多提供商 Agent 可在适配层将 Messages 响应转换成内部统一事件，再由 Agent 核心管理对话和工具执行。

## 3. 参考资料

- [Claude Platform 文档](https://platform.claude.com/docs/)
- [Messages API](https://platform.claude.com/docs/en/api/messages)
