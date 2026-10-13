# 响应报文

HTTP 响应报文由服务器发送，用于返回请求的处理结果及相关资源内容。

## 1. HTTP/1.1 报文结构

HTTP/1.1 响应报文由状态行、响应头、空行和可选的响应体组成。

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 16

{"name":"Alice"}
```

- [状态行](StatusLine.md)：包含 HTTP 版本、状态码和可选的原因短语，示例中分别为 `HTTP/1.1`、`200` 和 `OK`。
- [响应头](Headers.md)：携带内容格式、缓存指令等附加信息。
- [空行](BlankLine.md)：标记响应头结束。
- [响应体](Body.md)：携带返回的内容，并非所有响应都包含响应体。

## 2. 不包含响应体的情况

HEAD 请求的响应以及 1xx、204、304 响应不包含报文体。CONNECT 请求的成功响应在首部结束后切换为隧道通信，后续数据属于隧道，而不是 HTTP 响应体。

## 3. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 使用二进制帧承载响应，不采用上述文本报文格式。状态码通过 `:status` 伪首部表达，不传输原因短语；响应内容通过 DATA 帧传输。
