# PersistentVolume

PersistentVolume（PV，持久卷）是集群级资源，代表一份可供 Kubernetes 使用的存储，例如云盘或 NFS 目录。PV 可以由管理员预先创建，也可以由存储供给器按 PVC 的需求动态创建。

## 1. PV 配置

- [PersistentVolumeSpec](spec/Spec.md) 介绍容量、访问模式、回收策略，以及 CSI 和 Local 两种卷来源。
