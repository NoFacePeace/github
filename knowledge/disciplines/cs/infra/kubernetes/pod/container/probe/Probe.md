# Probe

Probe（探针）是 kubelet 用来检查容器状态的配置。三种探针都写在 Pod 的 `spec.containers` 对应条目下，可以分别配置。

## 1. 探针类型

| 探针 | 检查什么 | 连续失败达到阈值后 |
| --- | --- | --- |
| [Startup Probe](StartupProbe.md) | 应用是否完成启动。 | kubelet 终止容器，并按重启策略处理。配置后，存活和就绪探针等它首次成功才开始检查。 |
| [Readiness Probe](ReadinessProbe.md) | 应用当前是否能接收流量。 | 容器被标记为未就绪，Pod 通常随之退出对应 Service 的可用后端列表。 |
| [Liveness Probe](LivenessProbe.md) | 应用是否仍能正常运行。 | kubelet 终止容器，并按重启策略处理。 |

## 2. 检查方式与常用参数

kubelet 可以发 HTTP 请求（`httpGet`）、尝试建立 TCP 连接（`tcpSocket`）、在容器内运行命令（`exec`），或调用 gRPC 健康检查（`grpc`）。三篇探针笔记的示例都使用 HTTP，假设 `backend` 应用监听 `8080` 并实现对应的检查路径。HTTP 状态码 `200` 到 `399` 表示检查成功。

| 参数 | 含义 |
| --- | --- |
| `initialDelaySeconds` | 开始执行探针前的等待时间。 |
| `periodSeconds` | 检查间隔，单位为秒。 |
| `timeoutSeconds` | 单次检查的超时时间，单位为秒。 |
| `failureThreshold` | 判定探针失败所需的连续失败次数。 |
