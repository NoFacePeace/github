# 资源与对象

## 1. 资源类型与对象

Kubernetes API 提供 Pod、Service、EndpointSlice 等资源类型。名为 `backend` 的 Service 是 `Service` 类型的一个具体对象。日常所说的「创建一个 Service 资源」，指的就是创建一个 Service 对象。

## 2. 对象结构

下面用 `backend` Service 展示对象的基本结构。

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: default
  labels:
    app: backend
  annotations:
    example.com/description: "给前端提供 API"
spec:
  selector:
    app: backend
  ports:
    - name: http
      port: 80
      targetPort: 8080
```

| 字段 | 含义 |
| --- | --- |
| `apiVersion` | 对象使用的 API 组和版本。`v1` 表示核心 API 组的 `v1` 版本。 |
| `kind` | 对象的类型，例如 `Service`。 |
| `metadata` | 对象的身份和附加信息，例如名称、命名空间、标签和注解。 |
| [`spec`](Spec.md) | 期望的配置，例如 Service 要匹配哪些 Pod、提供哪个端口。 |
| `status` | 支持此字段的资源对象的当前状态，通常由 Kubernetes 或控制器维护。可用 `kubectl get service backend -n default -o yaml` 查看已创建的对象。 |

## 3. 元数据

- [Metadata](metadata/Metadata.md) 介绍对象的名称、命名空间、标签和注解。
