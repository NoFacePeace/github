# Volumes

`spec.volumes` 定义 Pod 可用的卷。每个条目指定卷名和数据来源，例如临时目录 `emptyDir`、ConfigMap 或持久卷声明。容器在自己的 [`volumeMounts`](../container/VolumeMounts.md) 中引用卷名，并用 `mountPath` 指定容器内的挂载路径。同一 Pod 中的多个容器可以挂载同一个卷。

## 1. 定义并挂载卷

下面是 `backend-pod` 的 `spec` 片段。`volumes` 定义名为 `cache` 的 `emptyDir` 卷，`backend` 容器将它挂载到 `/cache`。使用时将 `my-backend:1.0` 替换成实际镜像。

```yaml
volumes:
  - name: cache
    emptyDir: {}
containers:
  - name: backend
    image: my-backend:1.0
    volumeMounts:
      - name: cache
        mountPath: /cache
```

## 2. `emptyDir` 的生命周期

Pod 被分配到节点时，Kubernetes 创建一个初始为空的 `emptyDir` 卷。同一个 Pod 内的容器重启后，卷里的数据仍然保留。Pod 从节点移除时，卷和其中的数据会被删除。

## 3. 引用持久存储

需要在 Pod 重建后继续使用原来的数据时，可以用 `spec.volumes[*].persistentVolumeClaim` 引用同一命名空间里的 [PVC](../../storage/PersistentVolumeClaim.md)，再由容器按卷名挂载。
