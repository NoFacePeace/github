# VolumeMounts

`volumeMounts` 是容器的卷挂载列表，写在 `spec.containers` 中对应的容器条目下。`spec.volumes` 定义 Pod 可用的卷，容器通过 `volumeMounts` 指定要挂载的卷和容器内的路径。

## 1. 常用字段

| 字段 | 用途 |
| --- | --- |
| `name` | 要挂载的卷名，与同一个 Pod 中 `spec.volumes` 的某个条目名称一致。 |
| `mountPath` | 卷在容器内的挂载路径。 |
| `readOnly` | 设为 `true` 时，该容器通过这个挂载只读访问卷。 |
| `subPath` | 可选，挂载卷内的指定文件或子目录。省略时挂载整个卷。 |

## 2. 按名称挂载卷

假设 `backend-pod` 的 `spec.volumes` 中已经定义名为 `cache` 的 `emptyDir` 卷。在 `backend` 容器的配置中加入下面的片段，应用就可以在 `/cache` 读写这个卷。

```yaml
volumeMounts:
  - name: cache
    mountPath: /cache
```

同一 Pod 中的其他容器也可以引用 `cache`。每个容器分别指定自己的 `mountPath` 和 `readOnly`，因此同一个卷可以在不同容器中挂到不同路径。

## 3. 云盘连接与挂载

以把 EBS 云盘挂到 `backend` 容器的 `/data` 为例，连接和挂载是两个步骤。

1. **连接（attach）**。AWS 将云盘接到 Pod 所在的节点，节点可以访问对应的块设备。
2. **挂载（mount）**。节点上的 CSI 驱动准备卷的文件系统，kubelet 将卷交给容器运行时，由运行时把它挂到容器内 `mountPath` 指定的位置。

挂载完成后，应用通过 `/data` 读写云盘上的文件。
