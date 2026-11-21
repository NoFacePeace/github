# Liveness Probe

`livenessProbe` 检查运行中的应用是否仍能正常工作，例如应用是否卡死。探针连续失败达到 `failureThreshold` 时，kubelet 终止该容器，再按重启策略处理。已配置 `startupProbe` 时，存活检查会等到启动探针成功后开始。

## 1. 配置示例

把下面的配置放在 `spec.containers` 中 `backend` 的条目下。kubelet 定期请求 `/healthz`。应用卡死导致连续三次请求超时后，kubelet 会终止容器。

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3
```
