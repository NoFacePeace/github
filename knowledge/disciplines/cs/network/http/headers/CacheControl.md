# Cache-Control

`Cache-Control` 可用于请求和响应，通过逗号分隔的指令控制缓存行为。请求指令表达客户端对本次请求的缓存要求，响应指令约束响应的存储和复用。

## 1. 常见响应指令

| 指令 | 含义 |
| --- | --- |
| max-age=60 | 新鲜度期限为 60 秒，需扣除响应当前年龄 |
| s-maxage=300 | 为共享缓存指定新鲜度期限，覆盖 `max-age` 和 `Expires`，过期后必须验证 |
| no-cache | 允许存储，但每次复用前必须成功验证 |
| no-store | 禁止存储响应 |
| private | 只允许私有缓存存储 |
| public | 显式允许共享缓存存储，仍需满足其他规则 |
| must-revalidate | 过期后必须成功验证才能复用 |

这里的 `no-cache` 和 `private` 指不带字段名参数的形式。

## 2. 多个指令

```http
Cache-Control: private, max-age=60
```

表示只允许私有缓存存储，新鲜度期限为 60 秒。指令共同生效，不按出现顺序覆盖。

```http
Cache-Control: no-cache, max-age=3600
```

即使尚未过期，也必须先验证。`no-cache` 不等于禁止存储，禁止存储使用 `no-store`。

## 3. 请求指令

请求中的 `no-cache` 要求缓存先向源服务器验证，不能直接用已有响应满足请求。请求中的 `max-age=0` 限制可接受的响应年龄，并不与 `no-cache` 完全等价。
