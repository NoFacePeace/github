# 状态码

HTTP 状态码（Status Code）是响应中的三位数字，用于表示请求的处理结果。HTTP/1.x 将状态码放在状态行中，HTTP/2 与 HTTP/3 通过 `:status` 伪首部表达。

## 1. 分类

状态码的第一位表示类别。

| 类别 | 含义 |
| --- | --- |
| [1xx](1xx.md) | 信息响应，表示请求处理过程中的临时信息 |
| [2xx](2xx.md) | 成功，请求已成功接收、理解并接受 |
| [3xx](3xx.md) | 重定向，需要进一步操作才能完成请求 |
| [4xx](4xx.md) | 客户端错误，请求存在问题或无法被满足 |
| [5xx](5xx.md) | 服务器错误，服务器未能完成有效请求 |

## 2. 常见状态码

| 状态码 | 原因短语 | 含义 |
| --- | --- | --- |
| 100 | Continue | 客户端可以继续发送请求体 |
| 200 | OK | 请求成功，具体含义取决于请求方法 |
| 201 | Created | 请求成功并创建了一个或多个资源 |
| 204 | No Content | 请求成功，响应不包含报文体 |
| 301 | Moved Permanently | 目标资源已永久迁移；客户端可能将 POST 改为 GET |
| 302 | Found | 目标资源临时位于其他地址；客户端可能将 POST 改为 GET |
| 303 | See Other | 引导客户端通过检索请求访问另一个资源，通常使用 GET |
| 304 | Not Modified | 条件 GET 或 HEAD 请求的缓存表示仍可使用，响应不包含报文体 |
| 307 | Temporary Redirect | 临时重定向，重定向时保留请求方法 |
| 308 | Permanent Redirect | 永久重定向，重定向时保留请求方法 |
| 400 | Bad Request | 请求存在语法、分帧等客户端问题 |
| 401 | Unauthorized | 缺少有效身份认证凭据，通常需要认证 |
| 403 | Forbidden | 服务器理解请求，但拒绝执行 |
| 404 | Not Found | 未找到目标资源，或服务器不愿透露其存在 |
| 405 | Method Not Allowed | 目标资源不支持该请求方法 |
| 409 | Conflict | 请求与目标资源的当前状态冲突 |
| 429 | Too Many Requests | 请求频率过高 |
| 500 | Internal Server Error | 服务器遇到意外情况，无法完成请求 |
| 502 | Bad Gateway | 网关或代理从上游收到无效响应 |
| 503 | Service Unavailable | 服务暂时不可用，例如过载或维护 |
| 504 | Gateway Timeout | 网关或代理等待上游响应超时 |

## 3. 与原因短语的关系

状态码决定响应的协议语义，[原因短语](../messages/ReasonPhrase.md)仅供人阅读。客户端应依据状态码处理响应，而不是依赖 `OK` 等原因短语；HTTP/2 与 HTTP/3 不传输原因短语。
