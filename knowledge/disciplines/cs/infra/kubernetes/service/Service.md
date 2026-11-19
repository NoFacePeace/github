# Service

Service 为一组后端提供稳定的访问入口。Pod 重建后 IP 可能变化，集群内客户端可以继续用 Service 的 DNS 名称访问。

## 1. ClusterIP 与流量转发

`ClusterIP` 是默认类型。Service 的 DNS 名称解析为 Service 的虚拟 IP，集群的网络转发规则将连接送到一个可用后端。以下示例中，同一命名空间里的客户端可以访问 `http://backend:80`。

```text
backend:80 → Service 虚拟 IP:80 → 某个可用 Pod IP:8080
```

`port` 是客户端访问 Service 时使用的端口，`targetPort` 是后端接收流量的端口。`targetPort` 可以是数字，也可以引用 Pod 中声明的命名端口。

## 2. 通过标签选择 Pod

`spec.selector` 匹配同一命名空间中 Pod 的 `metadata.labels`。直接创建 Pod 时，在 `metadata.labels` 中声明标签。使用 Deployment 时，在 `spec.template.metadata.labels` 中声明，Deployment 创建的 Pod 会继承这些标签。

下面的 Service 选择带 `app: backend` 标签的 Pod。示例假定 `my-backend:1.0` 中的应用监听 TCP 8080，使用时替换为实际镜像。

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: default
spec:
  selector:
    app: backend
  ports:
    - name: http
      port: 80
      targetPort: 8080
---
apiVersion: v1
kind: Pod
metadata:
  name: backend-pod
  namespace: default
  labels:
    app: backend
spec:
  containers:
    - name: backend
      image: my-backend:1.0
```

## 3. 后端记录与服务发现

- [EndpointSlice](EndpointSlice.md) 说明后端如何记录，以及有 selector 和无 selector 时如何维护。Service 用 `spec.selector` 匹配 Pod 标签，EndpointSlice 用 `kubernetes.io/service-name` 关联同一命名空间里的 Service。
- [Headless Service](HeadlessService.md) 说明如何通过 Service DNS 名称发现后端 IP，以及如何直接连接 StatefulSet 的指定成员。
