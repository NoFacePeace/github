# Fetch API

Fetch API 是浏览器提供的 HTTP 请求接口，通过 `fetch()` 发起请求并获取 Response。它既可以读取完整响应体，也可以逐步读取响应流，支持配置请求方法、请求体以及浏览器允许设置的请求头。

## 1. 请求与响应

```javascript
const response = await fetch("/api/messages", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ text: "你好" }),
});

if (!response.ok) {
  throw new Error(`HTTP ${response.status}`);
}

const result = await response.json();
```

`fetch()` 在响应状态和首部可用时就可以返回 Response，不必等待响应体完整到达。HTTP 404、500 等状态通常不会让 fetch 的 Promise 自动拒绝，需要检查 `response.ok` 或状态码。

`response.text()` 和 `response.json()` 会读取完整响应体后再返回结果，因此不适合直接消费持续不断的事件流。

## 2. 逐步读取响应体

```javascript
const controller = new AbortController();
const response = await fetch("/events", { signal: controller.signal });

if (!response.ok || !response.body) {
  throw new Error(`无法读取响应流：HTTP ${response.status}`);
}

const reader = response.body.pipeThrough(new TextDecoderStream()).getReader();

try {
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    console.log(value);
  }
} finally {
  reader.releaseLock();
}

// 需要取消时，由按钮事件等调用 controller.abort()
```

`TextDecoderStream` 负责增量解码 UTF-8，处理字符跨字节块的情况。示例仅输出文本片段，并没有解析 SSE 事件或 NDJSON 记录。

正常结束时读取返回 `done: true`；异常中断时，读取可能抛错。使用 AbortController 可以取消请求及响应体消费；`releaseLock()` 只释放读取器锁，不等于取消请求。

能否逐步读取与 HTTP/1.1 是否使用 chunked 无关。即使服务端设置 Content-Length，已到达的数据仍然可以逐步读取；实际交付时间也受服务端和代理缓冲影响。

## 3. SSE 与 NDJSON 解析

原生 fetch 不会根据 Content-Type 自动把流转换成 SSE 事件或 NDJSON 记录，需要应用解析或使用适用的库：

- NDJSON：缓存文本片段，按换行符提取完整记录，再逐条 JSON 解析。
- SSE：按标准识别行结束和空行分隔，处理 `data`、`event`、`id`、`retry` 等字段；一个事件可能包含多行 data，不能只对每个读取块去掉 `data:` 前缀后解析。

网络分块与应用消息边界不一致，一次读取可能包含半条消息或多条消息。SSE 解析可以使用 eventsource-parser 等库；解析库本身不一定负责请求和重连，应检查具体能力。

## 4. 重连与恢复

fetch 不会像 EventSource 一样，在事件流结束后自动重新建立订阅，也不会维护 Last-Event-ID。应用需要决定是否重连、重试间隔、最后处理位置以及停止条件。

正常完成和异常断开都可能需要业务判断，不能把所有响应结束都当作失败重试。对于创建任务等 POST 请求，直接重试可能产生重复操作，应按接口约定使用任务标识或幂等机制恢复。

原生 EventSource 使用 GET，不能配置请求体或任意自定义请求头。需要通过 POST 或鉴权请求头接收 SSE 时，通常使用 fetch 或相应 SDK，并处理解析和恢复逻辑。

## 5. 与 WebSocket 的区别

浏览器 fetch 不支持通过 `ws://`、`wss://` 建立 WebSocket 连接，也不能仅靠设置 Upgrade 首部替代 WebSocket API。普通 HTTP 流式响应不具备 WebSocket 的消息帧和双向通信接口。

HTTP 客户端库可能额外提供 WebSocket 功能，但需要使用其专门的 WebSocket 接口。

## 6. 参考资料

- [MDN：使用 Fetch](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch)
- [MDN：Response.body](https://developer.mozilla.org/en-US/docs/Web/API/Response/body)
