# HTTPS

HTTPS（Hypertext Transfer Protocol Secure，安全超文本传输协议）是通过 [TLS](../tls/TLS.md) 保护的 HTTP 通信方式，可以简单理解为 **HTTPS = HTTP + TLS**。

## 1. HTTP 与 TLS 的分工

- **HTTP**：定义请求与响应的语义，例如请求方法、状态码、首部和正文。
- **TLS**（Transport Layer Security，传输层安全协议）：为通信提供加密、完整性校验和身份认证；浏览器通常通过验证服务器证书来确认服务器身份。

HTTPS 沿用 HTTP 的核心语义，因此在知识库中归在 HTTP 下面。TLS 是独立的安全协议，也可以保护其他应用协议。

## 2. 安全保障

- **加密**：防止旁观者直接读取传输中的密码、消息等内容。
- **完整性校验**：检测传输中的数据是否被篡改。
- **身份认证**：通过数字证书验证服务器身份与访问的域名是否匹配。

HTTPS 保护的是通信连接，不能保证网站本身可信；钓鱼网站也可以使用 HTTPS。

## 3. 协议栈与分层

| HTTP 版本 | 常见 HTTPS 协议栈 |
| --- | --- |
| HTTP/1.1、HTTP/2 | HTTP → TLS → TCP → IP |
| HTTP/3 | HTTP → QUIC → UDP → IP |

HTTP/3 使用 QUIC，QUIC 集成 TLS 1.3 的握手和安全机制，并由自身机制保护数据包。

TLS 没有严格对应 OSI 模型的某一层。对于基于 TCP 的 HTTPS，可以将它理解为位于 HTTP 与 TCP 之间；按功能通常类比为表示层，也涉及会话层功能。在 TCP/IP 模型中，通常将 TLS 归入应用层。

## 4. 默认端口

HTTP 默认端口为 **80**，HTTPS 默认端口为 **443**。HTTPS 使用 `https://` URL，例如 `https://example.com`。

现代 HTTPS 使用 TLS。SSL 是 TLS 的旧版前身，日常所说的“SSL 证书”通常仍指用于 TLS 的证书。
