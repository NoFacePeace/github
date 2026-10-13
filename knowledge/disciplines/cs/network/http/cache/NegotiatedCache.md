# 协商缓存

协商缓存指客户端通过条件请求向服务器验证已有缓存，服务器根据资源表示是否变化决定返回验证结果或最新内容。

## 1. 触发条件

缓存过期时通常需要验证；响应的 [Cache-Control](../headers/CacheControl.md) 包含不带字段名的 `no-cache` 时，每次复用前都必须验证，即使尚未过期。

**是否需要验证与使用什么验证，是两个不同的问题。**

- `max-age`、`Expires` 和 `no-cache` 等缓存规则决定是否需要验证。
- `ETag`、`Last-Modified` 提供需要验证时使用的依据，本身不会触发验证。

| 响应配置与缓存状态 | 下次复用时的行为 |
| --- | --- |
| `max-age=60` 和 `ETag`，缓存未过期 | 满足其他复用条件时直接使用，不发送请求 |
| `max-age=60` 和 `ETag`，缓存已过期 | 通常使用 ETag 发送条件请求验证 |
| `no-cache` 和 `ETag` | 无论是否过期，复用前都必须验证 |

因此，“缓存新鲜且允许直接复用”不仅要求未过期，还要求没有 `no-cache` 等必须先验证的限制。**同时配置有效期和验证标识，不代表每次都要验证。**

## 2. 验证字段

| 响应中的验证标识 | 后续验证请求携带 |
| --- | --- |
| [ETag](../headers/ETag.md) | [If-None-Match](../headers/IfNoneMatch.md) |
| [Last-Modified](../headers/LastModified.md) | [If-Modified-Since](../headers/IfModifiedSince.md) |

两个条件字段同时存在时，服务器优先判断 `If-None-Match`，忽略 `If-Modified-Since`。没有验证标识时，通常需要重新获取完整响应。

## 3. 请求示例

服务器首次返回：

```http
HTTP/1.1 200 OK
Cache-Control: max-age=60
ETag: "v1"
Content-Length: 5

Hello
```

缓存过期后，客户端发送：

```http
GET /greeting HTTP/1.1
Host: example.com
If-None-Match: "v1"
```

对于通常的条件 GET 请求：

- 表示未变化：返回 [304 Not Modified](../status-codes/304.md)，不发送内容；客户端复用已有内容并更新缓存信息。
- 表示已变化：同一次请求通常返回 `200 OK`、最新内容和新的验证标识，无须再发一次请求获取内容。

## 4. 与强缓存配合

同一响应可以同时包含有效期和验证标识：未过期时直接复用，过期后再验证。如果有 `no-cache`，则每次复用前都必须验证。
