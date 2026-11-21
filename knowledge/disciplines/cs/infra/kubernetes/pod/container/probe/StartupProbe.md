# Startup Probe

`startupProbe` 检查应用是否完成启动，适合需要较长时间初始化的容器。它首次成功后，kubelet 停止执行启动探针，并开始执行已配置的存活和就绪探针。连续失败达到 `failureThreshold` 时，kubelet 终止容器，再按重启策略处理。

## 1. 配置示例

把下面的配置放在 `spec.containers` 中 `backend` 的条目下。kubelet 每隔约 10 秒检查一次 `/startupz`，连续失败达到 30 次时判定启动失败。应用在完成启动后让该路径返回 `200`。

```yaml
startupProbe:
  httpGet:
    path: /startupz
    port: 8080
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 30
```
