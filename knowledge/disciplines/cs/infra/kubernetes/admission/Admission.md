# 准入控制

准入控制（Admission Control）是 Kubernetes API Server 处理请求的一环。以创建 Pod 为例，API Server 在完成身份认证和授权后执行准入检查，通过后才保存对象。

## 1. 请求处理顺序

```text
提交创建请求 → 身份认证 → 授权 → 变更准入 → 验证准入 → 保存对象
```

- **变更准入（Mutating Admission）** 可以修改请求对象，例如为 Pod 注入 sidecar。
- **验证准入（Validating Admission）** 检查变更后的对象是否符合要求，例如镜像是否来自允许的仓库。请求被拒绝时，API Server 会向客户端返回错误。

## 2. 通过 Webhook 扩展准入

- [Mutating Admission Webhook](MutatingAdmissionWebhook.md) 说明 API Server 如何调用外部服务获取补丁，以及如何将补丁应用到请求对象。
