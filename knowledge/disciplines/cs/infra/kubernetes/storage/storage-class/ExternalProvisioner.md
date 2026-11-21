# External Provisioner

`external-provisioner` 是负责 CSI 动态供给的常驻 Kubernetes 控制器。它通常以 `csi-provisioner` 容器运行，与 CSI 驱动的控制器端部署在同一个 Pod 里，并通过共享的 Unix socket 调用驱动的 CSI gRPC 接口。

## 1. 从 PVC 到 PV

以 `backend-data` PVC 选择 `fast-ssd` 并申请 `10Gi` 为例，动态供给分成以下几步。

1. `external-provisioner` 通过 Kubernetes API 监听 PVC，处理 StorageClass 的 `provisioner` 与它连接的驱动名称匹配、且需要动态供给的 PVC。
2. `external-provisioner` 发送 CSI `CreateVolume` 请求。EBS CSI 驱动调用 AWS API 创建实际的云盘，并返回卷 ID。
3. `external-provisioner` 调用 Kubernetes API 创建 PV 对象，把驱动名称和返回的卷 ID 写到 `spec.csi.driver` 和 `spec.csi.volumeHandle`。Kubernetes 内置的 PersistentVolume 控制器完成 PV 与 PVC 的绑定。

## 2. 何时处理 PVC

StorageClass 使用 `Immediate` 时，符合动态供给条件的 PVC 创建后即可开始供给。使用 `WaitForFirstConsumer` 时，`external-provisioner` 会等到 Pod 引用了 PVC，且调度器为动态供给选出节点，再结合该节点的拓扑请求驱动创建卷。以 EBS 为例，这让云盘能建在节点所在的可用区。
