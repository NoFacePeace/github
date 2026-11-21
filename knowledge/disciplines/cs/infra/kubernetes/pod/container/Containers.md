# Containers

`spec.containers` 是 Pod 的应用容器列表。每个条目对应一个容器，声明它的镜像及运行配置。同一 Pod 内的容器名称必须唯一。

## 1. 常用字段

| 字段 | 用途 |
| --- | --- |
| `name` | 容器的名称，例如 `backend`。 |
| `image` | 容器使用的镜像。 |
| `env` | 传给容器进程的环境变量。 |
| `resources.requests` | 调度时计入的资源需求，例如 CPU 和内存。 |
| `resources.limits` | 容器可使用的资源上限。 |
| [`volumeMounts`](VolumeMounts.md) | 按卷名指定挂载路径。 |

## 2. 配置示例

下面是 `backend-pod` 的 `spec` 片段，配置了一个名为 `backend` 的容器。使用时将 `my-backend:1.0` 替换成实际镜像。

```yaml
containers:
  - name: backend
    image: my-backend:1.0
    env:
      - name: APP_ENV
        value: "production"
    resources:
      requests:
        cpu: 250m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 256Mi
```

## 3. 辅助容器

- [Sidecar](Sidecar.md) 介绍和主容器一起运行的辅助容器，以及它们如何通过共享网络或卷协作。

## 4. 探针

- [Probe](probe/Probe.md) 介绍 kubelet 如何检查容器是否已启动、能否接流量，以及是否需要重启。
