# 持久存储

Kubernetes 通过 PersistentVolume（PV）和 PersistentVolumeClaim（PVC）管理可在 Pod 重建后继续使用的存储。以 `backend-pod` 把数据写到 `/data` 为例，容器内的挂载路径与底层存储之间有如下关系。

```text
容器 /data → Pod 中名为 data 的卷 → PVC backend-data → PV → 底层存储
```

## 1. PV 与 PVC

- [PersistentVolume（PV）](pv/PersistentVolume.md) 代表一份底层存储，例如云盘或 NFS 目录。
- [PersistentVolumeClaim（PVC）](PersistentVolumeClaim.md) 声明应用的存储需求。Pod 在自己的命名空间内引用 PVC，PVC 绑定到符合要求的 PV。

删除 Pod 后，保留 PVC 就可以让新 Pod 引用同一份存储。删除 PVC 后，绑定的 PV 及底层存储如何处理，由 PV 的回收策略决定。

## 2. 静态与动态供给

静态供给时，管理员先创建 PV，Kubernetes 再为 PVC 绑定符合要求的 PV。动态供给时，PVC 指定或使用默认的 [StorageClass](storage-class/StorageClass.md)。配置了动态供给器的 StorageClass 可以按需创建底层存储，并生成用于绑定的 PV。

## 3. 容器存储接口

- [CSI 接口](csi/CSI.md) 定义 Kubernetes 如何请求存储驱动创建、连接和挂载卷。
