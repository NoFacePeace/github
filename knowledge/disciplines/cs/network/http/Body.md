# 报文体

报文体（Message Body）用于承载 HTTP 请求或响应中的内容。请求中的报文体称为请求体，响应中的报文体称为响应体；并非所有报文都包含报文体。

## 1. 内容与格式

报文体可以包含 JSON、表单、文本或二进制数据等内容。`Content-Type` 首部说明内容的媒体类型。

```http
POST /users HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 16

{"name":"Alice"}
```

示例中的 `{"name":"Alice"}` 是请求体。

## 2. HTTP/1.1 中的边界

报文体位于首部后的空行之后，其长度或结束位置由报文分帧规则确定。常见方式包括 `Content-Length` 指定字节长度，以及 `Transfer-Encoding: chunked` 使用分块编码。部分响应通过关闭连接标记结束。

请求体不能通过关闭连接标记结束。HEAD 请求的响应以及 1xx、204、304 响应不包含报文体。

## 3. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 通过 DATA 帧传输报文内容，并通过流结束机制标记结束，不使用 HTTP/1.1 的分块传输编码。
