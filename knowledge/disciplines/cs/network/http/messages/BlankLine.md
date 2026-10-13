# 空行

在 HTTP/1.x 报文中，空行用于标记首部部分的结束，分隔首部和可选的[报文体](Body.md)。即使没有报文体，也需要这个空行。

## 1. 格式

每个首部字段行以回车换行（`CRLF`，即 `\r\n`）结束，随后再发送一个 `CRLF` 形成空行。因此，首部部分的结束通常表现为 `\r\n\r\n`。

```http
POST /users HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 16

{"name":"Alice"}
```

示例中，`Content-Length` 与 JSON 请求体之间的空行标记首部结束。报文体的长度或结束位置由报文分帧规则确定，不能仅通过空行判断。

## 2. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 使用二进制帧区分首部和数据，不使用空行分隔它们。
