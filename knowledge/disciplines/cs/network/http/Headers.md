# HTTP 首部

HTTP 首部（Headers）携带请求或响应的附加信息。客户端发送的首部称为请求头，服务器发送的首部称为响应头；部分字段可以用于两者。

## 1. HTTP/1.1 格式

每个首部字段由字段名、冒号和字段值组成，字段名不区分大小写。

首部部分以一个[空行](BlankLine.md)结束，随后可以是报文体。

```http
Host: example.com
Accept: application/json
Content-Type: application/json
```

## 2. 请求头

| 字段 | 用途 |
| --- | --- |
| Host | 指定目标主机及可选端口，HTTP/1.1 请求必须包含 |
| Accept | 指定客户端可接受的响应媒体类型 |
| Accept-Encoding | 指定客户端可接受的内容编码，如 gzip |
| Authorization | 携带身份认证凭据 |
| Cookie | 携带客户端保存的 Cookie |
| User-Agent | 描述客户端软件信息 |

## 3. 响应头

| 字段 | 用途 |
| --- | --- |
| Server | 描述服务器软件信息 |
| Set-Cookie | 指示客户端保存 Cookie |
| Location | 指定重定向目标或新建资源的位置 |
| WWW-Authenticate | 指定身份认证挑战 |

## 4. 请求与响应共用字段

| 字段 | 用途 |
| --- | --- |
| Content-Type | 说明报文内容的媒体类型 |
| Content-Length | 说明报文内容的字节长度；HEAD 等场景有特殊规则 |
| Cache-Control | 指定请求或响应的缓存指令 |
| Connection | 控制 HTTP/1.x 当前连接的使用，如 `close` |

## 5. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 通过二进制帧传输首部，字段名必须使用小写。它们使用 `:authority` 等伪首部表达请求控制信息，并禁止 `Connection` 等连接专用字段。
