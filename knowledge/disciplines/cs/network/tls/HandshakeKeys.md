# 握手密钥

握手密钥是 TLS 1.3 中用于加密握手记录的对称密钥，保护 ServerHello 之后的证书、签名和 Finished 等消息。

## 1. 派生方式

双方通过 ECDHE 得到共享秘密后，按 TLS 1.3 密钥调度生成 Handshake Secret，再结合握手记录哈希与方向标签派生握手流量秘密：

| 方向 | 流量秘密的派生标签 |
| --- | --- |
| 客户端 → 服务端 | `c hs traffic` |
| 服务端 → 客户端 | `s hs traffic` |

在不发生 HelloRetryRequest 的常规握手中，记录范围是 ClientHello 到 ServerHello。双方使用相同的基础秘密和记录，但方向标签不同，因此得到两个方向不同的流量秘密。

从每个方向的流量秘密继续使用 HKDF-Expand-Label，以 `key` 和 `iv` 为标签，分别派生实际加密密钥和 IV，长度由密码套件决定。

## 2. 如何使用

| 消息 | 使用的密钥 |
| --- | --- |
| 服务端 EncryptedExtensions、Certificate、CertificateVerify、Finished | 服务端发送方向的握手密钥 |
| 客户端 Finished，以及需要时的客户端证书和签名 | 客户端发送方向的握手密钥 |

双方都能派生两个方向的密钥。服务端发送密钥就是客户端接收密钥，客户端发送密钥就是服务端接收密钥。

## 3. 与 finished_key 的区别

finished_key 从对应方向的握手流量秘密单独派生，用于对握手记录哈希计算 HMAC，得到 Finished 内部的 verify_data。

握手记录加密密钥用于加密整个 Finished 消息，两者用途不同。后续 HTTP 等业务数据使用应用密钥保护。
