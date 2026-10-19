# WebSocket

WebSocket 是支持客户端与服务端通过持久连接交换消息的双向通信协议，可传输文本和二进制数据，常用于实时协作、聊天、游戏和持续双向数据传输。

## 1. 连接机制

经典 WebSocket 通过 HTTP/1.1 Upgrade 握手建立连接，成功后使用 WebSocket 帧传递消息，不再按普通 HTTP 请求与响应交换每条消息。HTTP/2 和 HTTP/3 也有相应的 WebSocket 建连机制。

地址通常以 `ws://` 或 `wss://` 开头，后者使用 TLS 加密。浏览器提供原生 WebSocket API。

## 2. 浏览器使用

```javascript
const socket = new WebSocket("wss://example.com/socket");

socket.onopen = () => {
  socket.send(JSON.stringify({ type: "message", text: "你好" }));
};

socket.onmessage = (event) => console.log(event.data);
socket.onerror = () => console.log("连接发生错误");
socket.onclose = (event) => console.log("连接关闭", event.code);

// 主动关闭
// socket.close();
```

示例地址需要替换为实际提供 WebSocket 服务的接口。服务端需使用支持 WebSocket 握手与帧处理的实现，不能仅将普通 HTTP 响应持续写入就视为 WebSocket。

## 3. 断线与恢复

浏览器原生 WebSocket 不自动重连，应用通常需实现退避重试、心跳或活性检测，以及连接恢复后的重新订阅。

WebSocket 本身不保存业务消息，也不自动补发离线期间的数据。需要可靠恢复时，应设计会话或任务标识、消息编号、确认位置和历史保留策略，并控制重复消息及重复操作。

## 4. 与 SSE 的选择

| 方面 | WebSocket | SSE |
|---|---|---|
| 通信方向 | 双向 | 当前事件流为服务端向客户端推送 |
| 数据形式 | 文本与二进制 | UTF-8 文本事件 |
| 浏览器接口 | WebSocket | EventSource，也可使用 fetch |
| 原生自动重连 | 不提供 | EventSource 提供 |
| 常见用途 | 频繁双向交互、实时协作 | 流式回答、进度和状态推送 |

如果主要需求是服务端推送，客户端偶尔通过 HTTP 提交操作，[SSE](SSE.md) 通常已经足够。选择方式应依据交互需求，而非仅依据是否需要实时更新。

## 5. 参考资料

- [RFC 6455：WebSocket 协议](https://www.rfc-editor.org/rfc/rfc6455)
- [MDN：WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
