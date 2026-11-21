# ShareProcessNamespace

`spec.shareProcessNamespace` 控制同一个 Pod 内的容器是否共享进程命名空间，默认值为 `false`。设置为 `true` 后，容器可以看到同一 Pod 内其他容器的进程。向其他容器的进程发送信号仍需满足 Linux 权限要求。

## 1. 配置示例

下面的 `backend-pod` 有两个容器。`backend` 运行 `sleep 3600`，`observer` 用来观察 Pod 内的进程。

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: backend-pod
spec:
  shareProcessNamespace: true
  containers:
    - name: backend
      image: busybox:1.36
      command: ["sleep", "3600"]
    - name: observer
      image: busybox:1.36
      command: ["sleep", "7200"]
```

## 2. 查看进程

Pod 启动后，可以在 `observer` 容器中查看进程列表。

```bash
kubectl exec backend-pod -c observer -- ps
```

输出中会看到 `backend` 的 `sleep 3600` 和 `observer` 自己的 `sleep 7200`。同一个 Pod 的容器也共享网络，可以通过 `localhost` 通信。网络共享独立于 `shareProcessNamespace` 的设置。
