# Content-Length

`Content-Length` 可用于 HTTP 请求和响应，字段值是十进制非负整数，通常表示报文内容的字节长度，而不是字符数量。

## 1. 示例

```http
Content-Type: application/json
Content-Length: 16

{"name":"Alice"}
```

示例中的 JSON 内容是 16 字节。计算长度时需要考虑字符编码；使用内容压缩时，长度对应压缩后的内容，不包含首部。

## 2. HTTP/1.1 报文边界

`Content-Length` 可以用于确定报文体的边界。发送方不得在同一报文中同时发送 `Content-Length` 和 `Transfer-Encoding`，以避免分帧歧义。

## 3. 特殊响应

- HEAD 响应可以携带该字段，值必须等于相同请求使用 GET 时会发送的内容长度，但 HEAD 响应没有报文体。
- 304 响应可以携带该字段，值必须等于相应请求返回 200 时会发送的内容长度，304 本身没有报文体。
- 1xx、204 响应及 CONNECT 的 2xx 响应不得携带该字段。

## 4. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 通过帧和流结束机制确定内容边界，仍可以携带 `Content-Length`。存在报文内容时，其声明长度必须与实际内容长度一致。
