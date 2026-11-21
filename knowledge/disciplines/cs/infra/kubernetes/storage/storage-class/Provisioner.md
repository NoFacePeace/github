# Provisioner

`provisioner` 是 StorageClass 的顶层字段，指定由哪个供给器为 PVC 准备底层存储并创建 PV。`kubernetes.io/no-provisioner` 表示通过预先创建的 PV 提供存储。

## 1. 常见配置

| 供给器或配置 | 供给方式 |
| --- | --- |
| `ebs.csi.aws.com` | AWS EBS CSI 动态供给。CSI 驱动创建云盘，配套的供给组件创建对应的 PV。 |
| `rancher.io/local-path` | Rancher Local Path Provisioner 动态供给。在带有 `WaitForFirstConsumer` 的配置中，供给器在为 Pod 选定的节点上创建目录，并创建对应的 PV。 |
| OpenEBS LocalPV LVM / ZFS | 通过对应的 CSI 驱动，从节点上预先配置的存储池中分配卷，并创建对应的 PV。 |
| `kubernetes.io/no-provisioner` | 静态供给。管理员预先准备存储和 PV，Kubernetes 为 PVC 绑定满足要求的 PV。内置 `local` PV 常用此配置。 |

## 2. 与 CSI 驱动和挂载的关系

以 `fast-ssd` 为例，StorageClass 的 `provisioner: ebs.csi.aws.com` 指定供给新卷的驱动。[External Provisioner](ExternalProvisioner.md) 说明如何请求驱动创建云盘、创建 PV 对象，再由内置控制器完成绑定。Pod 使用卷时，Kubernetes 根据已绑定 PV 中的 `spec.csi.driver` 和 `spec.csi.volumeHandle` 找到驱动和要挂载的卷。

对于预先创建的内置 `local` PV，卷来源写在 `spec.local` 下。Pod 调度到存储所在节点后，由该节点的 kubelet 准备本地卷。
