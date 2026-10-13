# 强缓存

强缓存指缓存响应仍新鲜且满足复用条件时，客户端直接使用缓存，不向服务器发送请求。这是常见的教学术语，HTTP 规范主要使用新鲜度和缓存复用等概念。

## 1. 有效期

服务器通过 [Cache-Control](headers/CacheControl.md) 等响应首部指定新鲜度：

```http
Cache-Control: max-age=60
```

表示新鲜度期限为 60 秒，复用时需扣除响应当前年龄，而不是每次收到响应后重新计时。

[Date](headers/Date.md) 表示响应生成时间，缓存结合它和 `Age` 等信息计算响应年龄；使用 `Expires` 时，新鲜度期限通常为 `Expires - Date`。

| 缓存类型 | 有效期依据，从高到低 |
| --- | --- |
| 私有缓存，如浏览器缓存 | `max-age` → [Expires](headers/Expires.md) |
| 共享缓存，如 CDN 缓存 | `s-maxage` → `max-age` → [Expires](headers/Expires.md) |

高优先级的有效指令覆盖低优先级字段，不取最早过期时间。没有显式有效期时，缓存可能按协议规则计算启发式有效期。

## 2. 复用限制

- 响应带有不带字段名的 `no-cache`：即使新鲜，复用前仍必须验证。
- 响应带有 `no-store`：不允许存储。
- 请求要求验证或缓存不匹配目标及 `Vary` 条件：不能仅凭有效期直接复用。

`ETag` 和 `Last-Modified` 本身不会触发验证。缓存过期后，通常可以使用这些标识进行协商缓存验证。
