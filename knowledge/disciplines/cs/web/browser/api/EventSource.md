# EventSource

EventSource 是浏览器提供的 Web API，用于建立 HTTP GET 连接并接收 [SSE](../../SSE.md) 事件流。它封装了事件解析、连接状态和自动重连，无需额外安装客户端库。

## 1. 基本使用

```javascript
const source = new EventSource("/events");

source.onopen = () => console.log("连接已建立");

source.onmessage = (event) => {
  console.log(event.lastEventId, event.data);
};

source.addEventListener("progress", (event) => {
  console.log(JSON.parse(event.data));
});

source.onerror = () => {
  console.log("连接状态", source.readyState);
};

// 不再需要接收时主动关闭，停止自动重连
// source.close();
```

`onmessage` 接收未指定事件类型或类型为 message 的事件。服务端使用 `event: progress` 等自定义类型时，需要通过 `addEventListener` 监听对应名称。`data` 是字符串，内容为 JSON 时由应用解析。

## 2. 连接状态与关闭

| readyState | 常量 | 含义 |
|---|---|---|
| 0 | EventSource.CONNECTING | 正在建立连接或等待重连 |
| 1 | EventSource.OPEN | 连接已建立 |
| 2 | EventSource.CLOSED | 已关闭，不再尝试重连 |

`error` 事件不一定意味着连接永久关闭。对于可重试的中断，浏览器会自动重连；如果希望继续接收，不应在每次 `onerror` 中直接调用 `close()`。

应用可以在收到自定义完成事件后调用 `close()`。服务端也可以通过 HTTP 204 响应告知客户端停止重连。

## 3. 断线与刷新恢复

同一个 EventSource 对象会记住服务端通过 `id` 设置的事件标识，自动重连时通过 `Last-Event-ID` 请求头发送该位置。服务端仍需自行实现事件存储和补发。

刷新页面或创建新对象后，不会自动继承旧对象的恢复位置。应用可保存任务标识和最后处理的事件编号，再通过自定义参数恢复：

```javascript
const source = new EventSource("/runs/run_123/events?after=42");
```

`after` 不是 API 内置参数，其解释方式由服务端约定。浏览器记录的接收位置也不等于业务处理确认；可靠消费仍需应用处理持久化与去重。

## 4. 限制与替代方式

- 原生接口使用 GET，不能配置 POST 请求体。
- 构造函数不提供任意自定义请求头选项，例如直接设置 Authorization 或首次连接的 Last-Event-ID。
- 不能通过同一连接向服务端持续发送消息，客户端可另发 HTTP 请求。
- 跨域请求需要服务端正确配置 CORS；需要跨域携带 Cookie 时，可使用 `new EventSource(url, { withCredentials: true })`，并由服务端配合凭据配置。
- 需要 POST、自定义请求头或自定义恢复策略时，可以使用 fetch 读取事件流或适用的客户端库，但需处理解析、取消和重连。

## 5. 参考资料

- [MDN：EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)
- [HTML 标准：Server-sent events](https://html.spec.whatwg.org/dev/server-sent-events.html)
