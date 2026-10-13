# 3xx 重定向

3xx 状态码表示需要进一步操作才能完成请求。部分状态码通过 `Location` 首部指定新的访问地址；304 用于缓存验证，不是跳转到另一个地址。

## 1. 常见状态码

| 状态码 | 原因短语 | 含义 |
| --- | --- | --- |
| 300 | Multiple Choices | 目标资源存在多个可选表示 |
| 301 | Moved Permanently | 永久重定向，客户端可能将 POST 改为 GET |
| 302 | Found | 临时重定向，客户端可能将 POST 改为 GET |
| 303 | See Other | 引导客户端通过检索请求访问另一个资源，通常使用 GET |
| [304](304.md) | Not Modified | 条件 GET 或 HEAD 请求的缓存表示仍可使用，不包含报文体 |
| 307 | Temporary Redirect | 临时重定向，自动重定向时保留请求方法和请求体 |
| 308 | Permanent Redirect | 永久重定向，自动重定向时保留请求方法和请求体 |
