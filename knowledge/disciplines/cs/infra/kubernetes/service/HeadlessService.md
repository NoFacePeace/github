# Headless Service

Headless Service 通过 `spec.clusterIP: None` 配置。集群 DNS 将 Service 名称解析为可用后端的 IP，客户端选择地址并直接连接后端端口。

## 1. 与普通 ClusterIP 的访问路径

两者都有稳定的 Service DNS 名称。差别在于 DNS 返回的地址，以及连接由谁选定后端。

| 访问方式 | DNS 返回 | 连接路径 |
| --- | --- | --- |
| 普通 ClusterIP | Service 的虚拟 IP | 客户端连接虚拟 IP，由集群转发到后端 |
| Headless | 可用后端的 IP | 客户端选定后端地址并直接连接 |

以下配置是普通 `backend` Service 的另一种定义。它通过相同的标签选择监听 TCP 8080 的 Pod。客户端查询 `backend` 的 A 或 AAAA 记录得到后端 IP，连接时使用后端端口 `8080`。

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: default
spec:
  clusterIP: None
  selector:
    app: backend
  ports:
    - name: http
      port: 8080
      targetPort: 8080
```

## 2. StatefulSet 的固定 Pod 域名

当名为 `backend` 的 StatefulSet 将 `spec.serviceName` 设为 `backend`，并与同命名空间的 Headless Service 搭配时，三个副本可分别通过 `backend-0.backend`、`backend-1.backend` 和 `backend-2.backend` 寻址。Pod 重建换 IP 后，成员的 DNS 名称保持稳定。默认情况下，Pod 就绪后才会发布相应的记录。需要在就绪前做成员发现时，可以在 Service 上设置 `spec.publishNotReadyAddresses: true`。

## 3. Headless 标记

在后端资源上看到 `service.kubernetes.io/headless` 时，可以把它理解为 Headless 标记。带有该标签的对象可通过标签存在性查询选中，标签值通常为空字符串。确认 Service 是否为 Headless 时，查看 `spec.clusterIP` 是否为 `None`。查找后端资源时，使用同一命名空间和 `kubernetes.io/service-name` 标签。
