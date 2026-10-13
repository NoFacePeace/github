# 4xx 客户端错误

4xx 状态码表示请求存在问题或无法被满足。排查时需要结合响应内容和服务约定，不能仅凭类别认定客户端实现存在缺陷。

## 1. 常见状态码

| 状态码 | 原因短语 | 含义 |
| --- | --- | --- |
| 400 | Bad Request | 请求存在语法、分帧等客户端问题 |
| 401 | Unauthorized | 缺少有效身份认证凭据，响应必须包含 `WWW-Authenticate` |
| 403 | Forbidden | 服务器理解请求，但拒绝执行 |
| 404 | Not Found | 未找到目标资源，或服务器不愿透露其存在 |
| 405 | Method Not Allowed | 目标资源不支持该方法，响应必须包含 `Allow` |
| 408 | Request Timeout | 服务器未在等待时间内收到完整请求 |
| 409 | Conflict | 请求与目标资源的当前状态冲突 |
| 410 | Gone | 目标资源已不可用，且这种情况可能是永久的 |
| 413 | Content Too Large | 请求内容超过服务器允许的大小 |
| 415 | Unsupported Media Type | 服务器不支持请求内容的格式 |
| 422 | Unprocessable Content | 内容类型和语法可理解，但无法处理其中的指令 |
| 429 | Too Many Requests | 请求频率过高，可以通过 `Retry-After` 提示重试时间 |
