# Mutating Admission Webhook

Mutating Admission Webhook（变更准入 Webhook）允许 API Server 在保存对象前调用外部 HTTPS 服务，并根据服务返回的补丁修改请求对象。下面以给 `backend` Pod 注入 sidecar 为例。

## 1. 配置调用和匹配规则

`MutatingWebhookConfiguration` 是集群级资源。`clientConfig` 指定 Webhook 的服务地址和用于校验服务证书的 CA，`rules` 和 `namespaceSelector` 决定哪些请求会触发调用。下面的配置匹配带有 `example.com/inject-sidecar: "true"` 标签的命名空间中的 Pod 创建请求。

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: sidecar-injector
webhooks:
  - name: sidecar-injector.example.com
    admissionReviewVersions: ["v1"]
    sideEffects: None
    failurePolicy: Fail
    clientConfig:
      service:
        namespace: webhook-system
        name: sidecar-injector
        path: /mutate
        port: 443
      caBundle: BASE64_CA_BUNDLE
    namespaceSelector:
      matchLabels:
        example.com/inject-sidecar: "true"
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
        scope: Namespaced
```

使用时需要部署对应的 HTTPS 服务，并将 `BASE64_CA_BUNDLE` 替换为对其 CA 证书的 PEM 内容进行 Base64 编码后的值。`failurePolicy: Fail` 表示调用失败时，API Server 会拒绝匹配的请求。

## 2. 用 JSON Patch 注入 sidecar

当用户提交包含 `backend` 容器的 Pod 时，API Server 把请求封装成 `AdmissionReview` 发送给 Webhook。Webhook 可以用下面的 JSON Patch 在 `spec.containers` 数组末尾添加一个容器。`example.invalid` 下的镜像地址是占位示例。

```json
[
  {
    "op": "add",
    "path": "/spec/containers/-",
    "value": {
      "name": "log-sidecar",
      "image": "example.invalid/platform/log-agent:1.0"
    }
  }
]
```

Webhook 返回 `AdmissionReview` 响应，`response.uid` 与请求的 UID 一致，`response.allowed` 为 `true`，`response.patchType` 为 `JSONPatch`，`response.patch` 包含 Base64 编码后的补丁。API Server 应用补丁并继续验证。通过后，集群中保存的 Pod 包含 `backend` 和 `log-sidecar`，可以通过 `kubectl get pod backend -o yaml` 查看。
