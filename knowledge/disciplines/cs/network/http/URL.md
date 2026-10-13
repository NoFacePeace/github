# URL

URL（Uniform Resource Locator，统一资源定位符）用于定位资源，说明通过什么方式、到哪里访问资源。URL 是 URI（Uniform Resource Identifier，统一资源标识符）的一种。

## 1. 组成

```text
https://example.com:443/users?page=1#profile
```

| 部分 | 示例 | 含义 |
| --- | --- | --- |
| 方案（Scheme） | `https` | 指定访问方式 |
| 主机（Host） | `example.com` | 目标主机的域名或 IP 地址 |
| 端口（Port） | `443` | 服务端口，可以省略默认端口 |
| 路径（Path） | `/users` | 资源路径 |
| 查询字符串（Query） | `page=1` | 查询信息，常用于传递参数 |
| 片段（Fragment） | `profile` | 标识资源内部位置，不随 HTTP 请求发送给服务器 |

## 2. 与 HTTP 请求的关系

直接向源站发送 HTTP/1.1 请求时，通常将路径和查询字符串放入请求行，将主机及必要的端口信息放入 `Host` 首部。例如，上述 URL 对应的请求可以是：

```http
GET /users?page=1 HTTP/1.1
Host: example.com
```

`https` 的默认端口是 `443`，因此示例中的 `Host` 可以省略端口。方案决定是否使用 TLS 等连接方式，片段 `#profile` 由客户端处理。
