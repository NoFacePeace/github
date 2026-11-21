# Readiness Probe

`readinessProbe` 检查应用当前是否能接收流量。探针连续失败达到 `failureThreshold` 后，kubelet 将容器标记为未就绪。常规 Service 会停止把新的流量转发给这个未就绪的 Pod。探针恢复成功，且 Pod 满足其他就绪条件后，Pod 可以重新接收流量。整个过程中容器可以继续运行。

## 1. 配置示例

把下面的配置放在 `spec.containers` 中 `backend` 的条目下。应用正在加载数据、暂时无法处理请求时，`/readyz` 可以返回 `503`，准备就绪后返回 `200`。

```yaml
readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 3
```
