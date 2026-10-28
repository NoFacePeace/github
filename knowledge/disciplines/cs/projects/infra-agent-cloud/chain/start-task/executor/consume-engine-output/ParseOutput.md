# 解析输出内容

Executor 调用所选引擎的解析方法，将标准输出行转换为平台需要的状态与事件数据。

## 1. 解析顺序

1. `parseSessionId`：提取引擎会话 ID，成功时更新内存中的 `cliSessionId`。
2. `parseMcpStatus`：提取引擎报告的 MCP 状态，成功时更新内存状态。
3. `parseProgress`：存在进度结果时，发出 `progress` 事件。
4. `parseChannelOutput`：遍历解析结果，逐条发出 `channel_output` 事件。

并非每一行都能解析出上述内容；没有相应结果时跳过。会话 ID 的提取在这里不直接更新 Session 表。

## 2. 存储操作

主要更新内存状态，事件通过输出管道传递，不直接读写 MySQL、Redis 或对象存储。

## 3. 下一步

解析出进度或通道输出时，[上报解析事件](ReportParsedEvents.md)。
