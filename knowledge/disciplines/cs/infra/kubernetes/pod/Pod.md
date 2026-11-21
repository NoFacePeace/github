# Pod

Pod 是 Kubernetes 中最小的可部署单元，包含一个或多个一起运行的容器。同一个 Pod 内的容器运行在同一个节点上，共享 Pod 的 IP 地址和端口空间，可以通过 `localhost` 通信。

## 1. Pod 配置

- [PodSpec](spec/PodSpec.md) 介绍 Pod 的期望配置，包括应用容器和卷。

## 2. 向容器提供 Pod 信息

- [Downward API](DownwardAPI.md) 说明如何通过环境变量或挂载文件，把 Pod 元数据和容器资源配置提供给应用。
