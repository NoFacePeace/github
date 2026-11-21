# PersistentVolumeSpec

PV 的 `spec` 描述这份存储的容量、访问方式、卷来源和回收策略。PV 与 PVC 绑定时，Kubernetes 会检查 PV 能否满足 PVC 对容量、访问模式、卷模式和存储类的要求。

## 1. 通用字段

| 字段 | 含义 |
| --- | --- |
| `capacity.storage` | PV 声明的存储容量，例如 `10Gi`。 |
| `accessModes` | 底层存储及其驱动支持的访问模式。 |
| `volumeMode` | `Filesystem` 将卷作为文件系统使用，`Block` 将卷作为裸块设备使用。 |
| `storageClassName` | PV 所属的存储类名称，用于与 PVC 的存储类匹配。 |
| `nodeAffinity` | 限定哪些节点可以使用该卷，例如卷所在的节点或可用区。 |

| 访问模式 | 含义 |
| --- | --- |
| `ReadWriteOnce`（RWO） | 可由单个节点读写挂载。同一节点上的多个 Pod 可以使用该卷。 |
| `ReadOnlyMany`（ROX） | 可由多个节点只读挂载。 |
| `ReadWriteMany`（RWX） | 可由多个节点读写挂载。 |
| `ReadWriteOncePod`（RWOP） | 可由集群中的单个 Pod 读写挂载，适用于支持该模式的 CSI 卷。 |

## 2. 回收策略

PVC 删除后，PV 的 `spec.persistentVolumeReclaimPolicy` 决定如何回收已释放的存储。

- `Retain` 保留 PV 和底层数据，由管理员手动处理。
- `Delete` 由支持该策略的存储插件删除 PV 和对应的底层存储。

## 3. 卷来源

每个 PV 声明一种卷来源。以下是两种常见配置。

- [CSI](CSI.md) 对应 `spec.csi`，指定管理卷的驱动及具体的卷 ID。
- [Local](Local.md) 对应 `spec.local`，指定某个节点上的本地磁盘或目录。
