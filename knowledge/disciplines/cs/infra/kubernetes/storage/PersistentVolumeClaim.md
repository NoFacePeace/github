# PersistentVolumeClaim

PersistentVolumeClaim（PVC，持久卷声明）是命名空间级资源，用于声明应用所需的容量和访问模式。Kubernetes 将 PVC 绑定到满足条件的 PV，或者通过指定的 StorageClass 动态供给存储。

## 1. 申请并绑定存储

下面的 `backend-data` PVC 申请 `10Gi` 存储，访问模式为 `ReadWriteOnce`。使用时将 `example-storage` 替换为集群中实际的 StorageClass 名称。选择已有 PV 时，容量、访问模式和存储类都需要符合申请条件。

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: backend-data
  namespace: default
spec:
  storageClassName: example-storage
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

## 2. 在 Pod 中引用 PVC

`backend-pod` 和 `backend-data` 都位于 `default` 命名空间。下面是 Pod `spec` 的片段，`claimName` 引用 PVC，`volumeMounts.name` 引用 Pod 中定义的 `data` 卷。应用通过 `/data` 访问持久存储。使用时将 `my-backend:1.0` 替换为实际镜像。

```yaml
volumes:
  - name: data
    persistentVolumeClaim:
      claimName: backend-data
containers:
  - name: backend
    image: my-backend:1.0
    volumeMounts:
      - name: data
        mountPath: /data
```

如果 StorageClass 的绑定模式是 `WaitForFirstConsumer`，PVC 会等到有 Pod 使用它时，再结合 Pod 的调度条件绑定或供给存储。

## 3. 查看绑定与引用

可以先创建引用 PVC 的 Pod。如果 `default` 命名空间里还没有 `backend-data` PVC，Pod 会等待调度，通常处于 `Pending`，调度事件会提示找不到该 PVC。创建符合要求的 PVC 后，调度器会重新尝试安排已有 Pod。

PVC 的 `spec.volumeName` 记录绑定的 PV 名称。Pod 对 PVC 的引用保存在 Pod 的 `spec.volumes[*].persistentVolumeClaim.claimName`。可以用下面的命令查看同一命名空间里引用该 PVC 的 Pod。

```bash
kubectl describe pvc backend-data -n default
```

输出中的 `Used By` 由 `kubectl` 根据 Pod 配置汇总。它表示哪些 Pod 声明引用了该 PVC，其中也可能包括正在等待调度的 Pod。
