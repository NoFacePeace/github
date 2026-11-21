# CSI

CSI（Container Storage Interface）是 Kubernetes 调用存储驱动的标准接口。PV 的 `spec.csi` 记录管理该卷的驱动，以及存储系统中的卷标识。

## 1. 驱动与卷标识

- `spec.csi.driver` 是已安装的 CSI 驱动名称，例如 `ebs.csi.aws.com`。
- `spec.csi.volumeHandle` 是驱动用来定位这份存储的标识，例如 EBS 云盘 ID。
- `spec.csi.fsType` 指定文件系统类型，例如 `ext4`。

接口和驱动的关系见 [CSI 接口](../../csi/CSI.md)。

## 2. 静态供给示例

下面的 `backend-data-pv` 将一块已创建的 `10Gi` AWS EBS 云盘注册为 PV，存储类为 `fast-ssd`。集群需要安装 AWS EBS CSI 驱动。使用时，将 `volumeHandle` 和 `nodeAffinity` 中的可用区替换为实际云盘的 ID 和可用区。

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: backend-data-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: fast-ssd
  csi:
    driver: ebs.csi.aws.com
    volumeHandle: vol-0123456789abcdef0
    fsType: ext4
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: topology.ebs.csi.aws.com/zone
              operator: In
              values:
                - ap-southeast-1a
```

创建 PV 后，`backend-data` PVC 可以选择 `fast-ssd`，申请 `10Gi` 和 `ReadWriteOnce`。Kubernetes 可以将符合条件的可用 PV 绑定给它。Pod 通过 `claimName: backend-data` 引用 PVC，PV 的 `nodeAffinity` 将 Pod 的调度范围限制在云盘所在的可用区。
