# SSE

SSE（Server-Sent Events，服务器发送事件）是服务器通过 HTTP 响应流持续向客户端推送文本事件的机制，常用于流式回答、任务进度和状态更新。

## 1. 基本机制

服务端返回 `Content-Type: text/event-stream`，以 UTF-8 文本持续写入事件，每个事件块以空行结束。浏览器原生 `EventSource` 使用 GET 请求接收事件；应用也可以通过 fetch 等 HTTP 客户端读取事件流。

```text
id: 42
event: progress
data: {"text":"正在查询订单"}

```

- `id`：事件标识，用于记录恢复位置。
- `event`：事件类型，省略时默认为 message。
- `data`：事件内容，可以承载 JSON 文本。
- `retry`：可选的重连间隔，单位为毫秒。
- 以冒号开头的行是注释，可用作心跳。

SSE 连接负责服务端到客户端的推送。客户端仍可通过独立的 POST 等请求发送消息或取消任务。

## 2. 浏览器使用

浏览器通过 [EventSource API](browser/api/EventSource.md) 接收 SSE，接口用法与限制见对应文档。

```javascript
const source = new EventSource("/events");

// 接收显式命名为 progress 的事件
source.addEventListener("progress", (event) => {
  console.log(event.lastEventId, JSON.parse(event.data));
});

// 接收未指定 event 字段或命名为 message 的事件
source.onmessage = (event) => console.log(event.data);

// 主动停止接收及自动重连
// source.close();
```

EventSource 自动解析事件流并在可重试的连接中断后尝试重连。使用 fetch 或普通 HTTP 客户端时，通常需要自行解析事件边界并实现重连；网络读取的数据块不一定对应完整事件。

## 3. 断线续传

同一个 EventSource 对象收到带 `id` 的事件后，会记住事件标识，自动重连时通过 `Last-Event-ID` 请求头发送恢复位置。服务端需要保存历史并实现补发，这个请求头本身不会自动恢复消息。

页面刷新后，需要应用保存并恢复订阅标识和事件位置。例如重新请求 `/events?after=42`；`after` 是应用约定，不是 SSE 标准字段。如果同时支持查询参数和请求头，可优先采用自动重连携带的请求头，避免使用旧的 URL 位置。

服务端应保证历史补发与实时订阅衔接期间不漏事件，客户端按标识去重。历史已过期时，应提供明确的恢复方式，例如先获取当前状态，再订阅后续事件。

## 4. 与 WebSocket 的选择

SSE 适合服务端持续推送文字事件；[WebSocket](WebSocket.md) 适合双方频繁发送消息或传输二进制数据。两者都需要应用设计持久化、去重和断线恢复。

部署 SSE 时还需处理反向代理缓冲、空闲超时和慢客户端，避免事件被延迟发送或连接长期占用资源。

## 5. 参考资料

- [HTML 标准：Server-sent events](https://html.spec.whatwg.org/dev/server-sent-events.html)
- [MDN：使用服务器发送事件](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)
