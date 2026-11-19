# EndpointSlice

Endpoint 是一个后端网络地址，例如 `10.244.1.7:8080`。EndpointSlice 是 `discovery.k8s.io/v1` API 对象，可以记录多个后端的地址、端口和就绪状态。一个 Service 可以对应多个 EndpointSlice。

## 1. 与 Service 的关联

EndpointSlice 通过自身的 `metadata.namespace` 和 `metadata.labels.kubernetes.io/service-name` 关联 Service。标签值等于 Service 的 `metadata.name`。例如，下列命令查询 `default` 命名空间中关联 `backend` Service 的所有 EndpointSlice。

```bash
kubectl get endpointslice -n default -l kubernetes.io/service-name=backend
```

EndpointSlice 可以作为 API 对象独立创建。匹配的 Service 存在后，集群才会将这些 EndpointSlice 用于该 Service 的后端发现。ClusterIP Service 的转发规则使用这些后端地址，Headless Service 的 DNS 使用这些地址生成后端记录。

## 2. 有 selector 时自动维护

Kubernetes 的 EndpointSlice 控制器根据 Service 的 `spec.selector` 查找同一命名空间内标签匹配的 Pod，自动创建 EndpointSlice，并添加 `kubernetes.io/service-name` 标签。Pod 增减、IP 或就绪状态变化时，控制器会更新相应的记录。

## 3. 无 selector 时手动维护

省略 Service 的 `spec.selector` 后，可以手动创建 EndpointSlice，也可以由其他控制器维护。假设 `default/backend` 已配置为无 selector，且 Service 端口名为 `http`，下面的 EndpointSlice 指定一个后端。将示例 IP 替换为实际可达的后端地址。

```yaml
apiVersion: discovery.k8s.io/v1
kind: EndpointSlice
metadata:
  name: backend-manual
  namespace: default
  labels:
    kubernetes.io/service-name: backend
    endpointslice.kubernetes.io/managed-by: manual
addressType: IPv4
ports:
  - name: http
    protocol: TCP
    port: 8080
endpoints:
  - addresses: ["10.244.1.7"]
    conditions:
      ready: true
```

EndpointSlice 的端口名称和协议应与 Service 对应端口一致，端口号填写后端实际监听的端口。后端 IP 或就绪状态变化时，由 EndpointSlice 的维护者更新记录。这种配置也可用于集群外可达的后端。

## 4. 旧版 Endpoints

`Endpoints` 是旧版 API 对象，一项 Service 的后端集中记录在一个对象中。EndpointSlice 支持分片，后端变动时可以只更新受影响的片。排查 Service 后端时优先查询 EndpointSlice。
