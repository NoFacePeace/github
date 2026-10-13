# 5xx 服务器错误

5xx 状态码表示服务器发生错误或无法完成请求。是否重试需要结合具体错误、请求方法的幂等性及服务约定判断。

## 1. 常见状态码

| 状态码 | 原因短语 | 含义 |
| --- | --- | --- |
| 500 | Internal Server Error | 服务器遇到意外情况，无法完成请求 |
| 501 | Not Implemented | 服务器不支持完成请求所需的功能 |
| 502 | Bad Gateway | 网关或代理从上游收到无效响应 |
| 503 | Service Unavailable | 服务暂时不可用，例如过载或维护，可以通过 `Retry-After` 提示重试时间 |
| 504 | Gateway Timeout | 网关或代理等待上游响应超时 |
| 505 | HTTP Version Not Supported | 服务器不支持请求所使用的 HTTP 主版本 |
