# 请求报文

HTTP 请求报文由客户端发送，用于向服务器请求资源或提交数据。

## 1. HTTP/1.1 报文结构

HTTP/1.1 请求报文由请求行、请求头、空行和可选的请求体组成。

```http
POST /users HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 16

{"name":"Alice"}
```

- [请求行](RequestLine.md)：包含请求方法、请求目标和 HTTP 版本。
- [请求头](Headers.md)：携带目标主机、请求体格式等信息。
- [空行](BlankLine.md)：分隔请求头和请求体。
- [请求体](Body.md)：携带提交的数据，并非所有请求都包含请求体。

## 2. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 使用二进制帧承载请求，不采用上述文本报文格式。请求方法、目标等信息通过伪首部（如 `:method`、`:path`）表达，请求体通过 DATA 帧传输。
