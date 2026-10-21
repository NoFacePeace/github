# OpenAI API

OpenAI API 包含多种接口格式。Chat Completions 与 Responses 都可用于调用模型，但请求结构、工具调用和流式事件不同；不能只替换 URL 就认为完成迁移。

## 1. 模型接口

- [Chat Completions](ChatCompletions.md)：以消息列表组织对话，常见于第三方 OpenAI 兼容服务。
- [Responses](Responses.md)：以不同类型的 Item 表示消息、工具调用及结果，支持响应链等上下文管理方式。

## 2. 接口选择与兼容

希望接入多家模型服务时，可优先考察 Chat Completions 兼容接口；使用 OpenAI 功能时，再根据需要选择 Responses。兼容某个端点不代表完整支持工具调用、图片、结构化输出、推理参数和流式事件，需要按服务商与模型确认。

多提供商 Agent 可以在适配层将各家的响应转换成内部统一事件。Agent 核心负责消息历史与工具执行，适配层负责协议解析、参数分片聚合和字段转换。

## 3. 参考资料

- [OpenAI API 文档](https://developers.openai.com/api/docs/)
- [迁移到 Responses API](https://developers.openai.com/api/docs/guides/migrate-to-responses)
