# StorageClass

StorageClass（存储类）定义 Kubernetes 如何供给某类存储，包括使用哪个存储插件、创建什么类型的存储，以及何时绑定或供给卷。它是独立的集群级资源。PVC 通过 `spec.storageClassName` 引用存储类的名称。

## 1. 定义和引用存储类

下面的 `fast-ssd` 示例适用于已安装 AWS EBS CSI 驱动的集群。`provisioner`、`parameters`、`reclaimPolicy` 和 `volumeBindingMode` 都写在 StorageClass 的顶层。

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

[`provisioner`](Provisioner.md) 选择存储供给器，`parameters` 指定该供给器支持的参数。这里 `type: gp3` 表示创建 AWS EBS gp3 SSD 云盘。`reclaimPolicy` 会应用到动态创建的 PV。这里指定 `Delete`，因此 PVC 删除并释放 PV 后，PV 和对应的云盘都会被删除。

例如，`backend-data` PVC 可以申请 `10Gi` 并设置 `spec.storageClassName: fast-ssd`。可以用下面的命令查询这份存储类的完整配置。

```bash
kubectl get storageclass fast-ssd -o yaml
```

## 2. 绑定和供给时机

| `volumeBindingMode` | 行为 |
| --- | --- |
| `Immediate` | 默认值。PVC 创建后就尝试绑定已有 PV，或按需供给新卷。 |
| `WaitForFirstConsumer` | 等到有 Pod 引用 PVC，再结合该 Pod 的调度条件选择已有 PV，或按需供给新卷。 |

例如，`backend-pod` 只能调度到可用区 A，而 EBS 云盘也必须和节点在同一可用区。提交引用 PVC 的 Pod 后，调度器先为它选定合适的节点。如果需要新存储，供给器就在该节点所在的可用区创建云盘并生成 PV。PVC 绑定后，Pod 才能挂载云盘并启动。
