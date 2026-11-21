# CSI 驱动

CSI 驱动是实现 Container Storage Interface 的程序，负责操作具体的存储系统。以 AWS EBS CSI 驱动 `ebs.csi.aws.com` 为例，它可以调用 AWS API 创建云盘，也可以在节点上准备卷供 Pod 使用。

## 1. 控制器端与节点端

| 组件 | 运行位置与职责 |
| --- | --- |
| 控制器端 | EBS 驱动的控制器端通常以 Deployment 运行，处理创建云盘、将云盘连接到节点等操作。 |
| 节点端 | EBS 驱动的节点端通常以 DaemonSet 运行，在需要使用 EBS 卷的节点上准备和挂载卷。 |

CSI 定义了 gRPC 方法和请求格式。调用方通过 Unix socket 向驱动进程发送请求。

## 2. Kubernetes 如何找到驱动

**动态供给时**，`external-provisioner` 已通过共享的 Unix socket 连接控制器端驱动。它读取驱动报告的名称，再处理 StorageClass 的 `provisioner` 与该名称匹配、且需要动态供给的 PVC。它通过 CSI `CreateVolume` 请求让驱动创建底层存储。

**节点挂载时**，`node-driver-registrar` 向所在节点的 kubelet 注册驱动名称和 CSI socket 路径。kubelet 从 PV 的 `spec.csi.driver` 读取驱动名称，找到已注册的 socket，再发送 `NodePublishVolume` 等 CSI 请求。请求中的卷 ID 来自 PV 的 `spec.csi.volumeHandle`。

`CSIDriver` 是记录驱动行为和能力的 Kubernetes API 对象，例如通过 `spec.attachRequired` 声明是否需要先连接卷。节点上的驱动进程通过已注册的 CSI socket 接收请求。
