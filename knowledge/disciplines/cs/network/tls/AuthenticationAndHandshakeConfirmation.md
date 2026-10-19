# 身份认证与握手确认

这是 TLS 1.3 握手的第二阶段。在双方派生握手密钥后，通过证书、签名和 Finished 认证身份、验证握手完整性并确认密钥材料。以下以仅认证服务器的常规完整握手为例。

## 1. 服务端发送认证消息

身份凭证及其验证依据见[证书](certificates/Certificate.md)；签名机制见[数字签名](../../cryptography/DigitalSignature.md)。

服务端在 ServerHello 后连续发送以下消息，使用服务端发送方向的握手流量密钥加密：

| 消息 | 作用 |
| --- | --- |
| EncryptedExtensions | 返回其他协商参数 |
| Certificate | 提供服务器证书 |
| CertificateVerify | 对握手上下文签名，证明持有证书对应私钥 |
| Finished | 验证握手记录完整性，确认密钥材料 |

客户端验证证书信任链、有效期、访问域名等，并验证 CertificateVerify 的签名和 Finished 的校验值。证书主要用于身份认证，不用于逐条加密应用数据。

## 2. 客户端发送 Finished

客户端验证成功后，发送自己的 Finished，使用客户端发送方向的握手流量密钥加密。服务端收到并验证后，服务端一侧的握手完成。

Finished 内部携带 verify_data 校验值，不携带 HTTP 等应用数据：

```text
该方向的握手流量秘密 → 派生 finished_key
                                 ↓
              对此前握手记录哈希计算 HMAC
                                 ↓
                     得到 Finished 校验值
                                 ↓
                  用该方向握手流量密钥加密消息
```

计算校验值的 finished_key 与加密整条消息的握手流量密钥用途不同。

## 3. 1-RTT 与应用数据发送时机

客户端发送 ClientHello 后，等待服务端整组消息返回，约需一个网络往返。验证完成后，客户端可以连续发送 Finished 和应用数据，无需等待服务端再次确认。

因此，客户端 Finished 仍属于握手，但不额外引入一个往返等待。这就是常规 TLS 1.3 握手称为 1-RTT 握手的含义，不包含 TCP 建连时间。

## 4. 下一阶段

结合到服务端 Finished 为止的握手记录，可以派生初始应用流量秘密及其加密密钥，进入[应用数据传输](ApplicationDataTransfer.md)。
