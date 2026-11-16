# 自动重连

EventSource 的自动重连是指：事件流连接发生可重试的中断后，浏览器等待一段时间，再自动发起 HTTP GET 请求，建立新连接并继续接收事件。它不是恢复原来的连接，而是替应用执行重新请求的逻辑。

## 1. 触发与过程

已建立的事件流可能因网络中断、代理空闲超时或服务端结束响应而断开。对于可重试的断开，即使服务端正常结束了响应，浏览器也会尝试重连：

```text
GET /events → 接收事件 → 响应结束或连接中断
                              ↓
                     等待重连间隔
                              ↓
                     再次 GET /events → 继续接收事件
```

重连期间，`readyState` 为 `EventSource.CONNECTING`；重新连接成功后变为 `EventSource.OPEN`，并再次触发 `open` 事件。`error` 事件可能只是表示进入重连流程，不一定表示永久失败。

```javascript
const source = new EventSource("/events");

source.onopen = () => console.log("首次连接或重连成功");
source.onmessage = (event) => console.log(event.data);
source.onerror = () => {
  if (source.readyState === EventSource.CONNECTING) {
    console.log("连接中断，等待浏览器自动重连");
  }
};
```

## 2. 重连间隔

默认重连间隔由浏览器决定。服务端可以在事件流中发送 `retry` 字段，设置后续重连的等待时间，单位为毫秒：

```text
retry: 3000

```

这里表示重连前等待约 3 秒。字段值必须仅由十进制数字组成，否则会被忽略；浏览器也可能因退避策略或网络不可用而延长实际等待时间。`retry` 是事件流字段，不是 HTTP 响应头。

## 3. 重连与续传

自动重连只负责重新建立连接，不保证补上断线期间的消息。服务端可以通过 `id` 为事件设置恢复位置：

```text
id: 42
data: {"percent":80}

```

同一个 EventSource 对象记住该位置后，会在自动重连的请求中携带：

```http
GET /events HTTP/1.1
Last-Event-ID: 42
```

服务端需要读取 `Last-Event-ID`，查找历史并补发后续事件。仅发送 `id` 不会自动实现续传，应用仍需处理消息去重和历史过期。

## 4. 停止重连

- 客户端调用 `source.close()`，会关闭连接并停止重连。
- 服务端返回 HTTP 204，告知浏览器停止重连。
- 不可重试的失败，例如响应的 Content-Type 不是 `text/event-stream`，也会使连接进入 `EventSource.CLOSED` 状态。

服务端仅结束当前事件流响应，通常会触发重连。对于有明确终点的任务，可以发送自定义完成事件，由客户端在处理后调用 `close()`。

## 5. 与 fetch 的区别

原生 fetch 可以读取流式响应，但不会自动重连，也不会自动解析 SSE 或维护 `Last-Event-ID`。响应正常结束时，流读取返回 `done: true`；异常中断时，读取可能抛错。应用需要自行决定是否重新请求，并实现重试间隔、恢复位置和停止条件，或使用客户端库处理。
