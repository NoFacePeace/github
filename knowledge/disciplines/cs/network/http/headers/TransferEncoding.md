# Transfer-Encoding

`Transfer-Encoding` 是 HTTP/1.1 中描述报文体传输编码的首部。常见取值 `chunked` 将报文体编码为若干块，在无需提前知道总长度的情况下，明确标记报文体的结束。

## 1. chunked 格式

```http
Transfer-Encoding: chunked
```

每个普通数据块由“十六进制字节长度 + CRLF + 数据 + CRLF”组成。最后发送长度为 `0` 的块，再结束可选的尾部字段区域；没有尾部字段时以一个空行结束。

例如，要发送 `hello world`，编码后的响应体可以表示为：

```text
5\r\nhello\r\n6\r\n world\r\n0\r\n\r\n
```

示例中的 `\r\n` 表示实际的回车换行字节，并非字面文本。`5` 表示 `hello` 有 5 字节，`6` 表示 ` world` 有 6 字节，`0` 表示不再有数据块。完整报文还需接收完尾部字段区域的结束空行。

## 2. 与 Content-Length 的区别

- `Content-Length`：提前声明完整内容的字节长度，接收方收够后知道报文体结束。
- `chunked`：逐块声明长度，最后通过结束标记结束报文体，适合总长度未知的内容。

发送方不得在同一报文中同时发送 `Content-Length` 和 `Transfer-Encoding`。使用 chunked 后可以结束当前响应而保持连接打开，供后续请求复用。

没有使用 chunked 不代表必须收完整个响应才能读取。只要客户端接口支持增量读取，已到达的内容就可以被消费；Content-Length 同样可以用于流式传输。

## 3. 与应用格式和数据读取的区别

`Content-Type` 说明内容是什么，例如 SSE 的 `text/event-stream` 或 NDJSON 的 `application/x-ndjson`；`Transfer-Encoding` 说明 HTTP/1.1 如何编码和界定传输的报文体。

两者可以同时使用，但一个 HTTP chunk 不一定对应一条 SSE 事件或一条 NDJSON 记录。HTTP 客户端通常自动去掉 chunked 标记，应用看到的是解码后的内容流；读取接口返回的数据块也不一定对应原始 HTTP chunk。

chunked 不是开启流式传输的开关，不保证框架或代理会立即转发每一块数据。框架通常自动选择是否使用该编码，应用不需要手动生成块长度。

## 4. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 使用各自的帧和流结束机制，不使用 `Transfer-Encoding: chunked`，也不得发送该首部。它们同样支持长度已知或未知的流式内容。

## 5. 参考资料

- [RFC 9112：Chunked Transfer Coding](https://www.rfc-editor.org/rfc/rfc9112.html#section-7.1)
