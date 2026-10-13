# 请求目标

请求目标（Request Target）是 HTTP/1.x 请求行中位于请求方法与 HTTP 版本之间的部分，用于指定请求的目标。

请求目标之后是 [HTTP 版本](Version.md)字段，例如 `HTTP/1.1`。

## 1. 常见形式

```http
GET /users?page=1 HTTP/1.1
```

示例中的 `/users?page=1` 是请求目标，包含路径 `/users` 和查询字符串 `page=1`，不包含 URL 的片段部分。

## 2. 四种形式

| 形式 | 示例 | 使用场景 |
| --- | --- | --- |
| 源站形式（origin-form） | `/users?page=1` | 直接向源站发送请求，CONNECT 除外；OPTIONS 也可使用星号形式 |
| 绝对形式（absolute-form） | `http://example.com/users?page=1` | 向 HTTP 代理发送请求，CONNECT 和服务器级 OPTIONS 除外 |
| 权限形式（authority-form） | `example.com:443` | CONNECT 请求，指定隧道目标的主机和端口 |
| 星号形式（asterisk-form） | `*` | 服务器级 OPTIONS 请求，查询整个服务器的通信选项 |

源站形式中的路径不能为空；URL 没有路径时，使用 `/`。

## 3. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 没有文本请求行，通常通过 `:scheme`、`:authority` 和 `:path` 等伪首部表达目标信息。普通 CONNECT 请求使用 `:authority` 指定主机和端口，不包含 `:scheme` 和 `:path`。
