# 领取执行凭证

Runner 携带 Job 的授权引用、执行标识、租约及 Runner 实例身份，向 Server 请求本轮执行所需的敏感输入。

## 1. 服务端校验

Server 根据 [SecretGrant](../../../../storage/mysql/secret-grant/SecretGrant.md) 校验调用身份、授权绑定、领取截止时间及有效执行租约，再解密并交付适用的模型凭证、环境变量与资源准备敏感信息。

Runner 不直接读取授权表，也不应将领取结果输出到普通日志。

## 2. 后续步骤

领取后检查是否已有停止请求或服务关闭要求，必要时回传停止并结束准备；否则进入[启动租约续期](StartLeaseRenewal.md)。
