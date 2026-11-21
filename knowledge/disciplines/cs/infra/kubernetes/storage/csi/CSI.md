# CSI 接口

CSI（Container Storage Interface，容器存储接口）是一套标准 gRPC 接口。Kubernetes 的存储组件通过 CSI 向驱动发出请求，驱动再对接实际的存储系统，例如 AWS EBS。不同存储系统的驱动可以实现同一套接口。

## 1. 接口与驱动

CSI 规定调用方法、参数和返回值。例如，调用方用 `CreateVolume` 请求创建卷，驱动返回具体的卷 ID。驱动如何部署、注册以及接收请求，见 [CSI 驱动](Driver.md)。

## 2. 在 Kubernetes 中怎么用

以 AWS EBS 为例，StorageClass 的 `provisioner: ebs.csi.aws.com` 指定动态供给使用的驱动。PV 的 `spec.csi.driver` 记录管理这份卷的驱动，`spec.csi.volumeHandle` 记录具体的 EBS 云盘 ID。

| 阶段 | CSI 请求示例 |
| --- | --- |
| 供给 | `CreateVolume` 请求驱动创建实际的卷。 |
| 连接 | `ControllerPublishVolume` 请求驱动将需要连接的卷接到指定节点。 |
| 挂载 | `NodePublishVolume` 请求节点上的驱动把卷提供给 Pod。 |
