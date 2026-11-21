# Sidecar

Sidecar（边车容器）是在同一个 Pod 中和主容器一起运行、负责辅助工作的容器。常见职责包括收集日志、转发流量或协调存储供给。

## 1. 容器如何协作

同一个 Pod 内的容器共享网络，可以通过 `localhost` 通信。两个容器也可以各自挂载同一个 Pod 卷，用它交换文件或通过 Unix socket 通信。

Sidecar 这个名称描述容器的辅助角色。辅助容器可以写在 `spec.containers` 下。Kubernetes 也支持原生 sidecar，通过在 `spec.initContainers` 中为该容器设置 `restartPolicy: Always`，让它按初始化顺序启动，并在 Pod 运行期间持续运行。

## 2. EBS CSI 示例

EBS CSI 控制器 Pod 中，EBS 驱动容器和 `csi-provisioner` 辅助容器挂载同一个卷。驱动在其中监听 Unix socket，`csi-provisioner` 通过这个 socket 发送 `CreateVolume` 请求。驱动创建云盘并返回卷 ID 后，`csi-provisioner` 调用 Kubernetes API 创建对应的 PV。
