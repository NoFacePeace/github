# 上报解析事件

Executor 将解析得到的进度、回复等数据交给 `emit`，通过自身标准输出发送协议事件。

## 1. 事件封装

统一补充 `schemaVersion`、`executionId`、递增的 `seq`、时间戳、事件类型与适用的 `payload`。通过同一写入队列发送，保证本轮事件顺序。

CLI 将每个事件序列化为一行 JSON，写入 `process.stdout`；写入缓冲区已满时，异步等待 `drain` 后继续。

## 2. 通用示例

```json
{"schemaVersion":1,"executionId":"exec-123","seq":8,"timestamp":"2026-10-28T08:00:00.000Z","type":"progress","payload":{}}
```

此示例只展示事件外壳，实际 payload 取决于引擎解析结果。原始输出事件也使用相同封装与发送通道。

## 3. 下一步

协议事件由[Runtime 接收并转发](ForwardRuntimeEvents.md)。这是同一条事件传输链，Executor 不直接调用 Server 接口。
