# Local

`spec.local` 是 Kubernetes 内置的卷来源。`spec.local.path` 指向节点上已准备好的磁盘、分区或目录。Local PV 必须配置 `spec.nodeAffinity`，让调度器知道卷位于哪个节点。

## 1. 准备存储类和 Local PV

管理员先在 `node-a` 上准备 `/mnt/backend-data`，再创建 `local-storage` StorageClass 和 `backend-local-pv`。示例中的节点名应替换为实际节点的 `kubernetes.io/hostname` 标签值，容量也应与准备的存储相符。

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: backend-local-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt/backend-data
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - node-a
```

`kubernetes.io/no-provisioner` 表示通过预先准备的 PV 提供存储。`WaitForFirstConsumer` 让 Kubernetes 等到有 Pod 引用 PVC，再结合 Pod 的调度条件选择匹配的本地卷。

## 2. 绑定和挂载

使用者创建 `backend-data` PVC，选择 `local-storage`，申请 `10Gi` 和 `ReadWriteOnce`。随后创建 `backend-pod`，在 `spec.volumes` 中通过 `claimName: backend-data` 引用 PVC，并在容器中将对应的卷挂到 `/data`。

Kubernetes 结合 Pod 的调度条件和 PV 的节点亲和性完成绑定及调度。Pod 调度到 `node-a` 后，由该节点的 kubelet 准备本地卷，容器运行时将其挂到 `/data`。数据保留在原节点上。节点不可用时，其他节点上的 Pod 无法挂载这份数据。
