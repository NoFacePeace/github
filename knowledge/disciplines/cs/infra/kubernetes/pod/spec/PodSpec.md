# PodSpec

`PodSpec` 是 Kubernetes API 中 Pod 的 `spec` 字段对应的类型，描述 Pod 要运行哪些容器、提供哪些卷等期望配置。直接声明 Pod 时写在 `spec` 下，通过 Deployment 创建 Pod 时写在 `spec.template.spec` 下。

## 1. 应用容器

- [Containers](../container/Containers.md) 介绍 `spec.containers`，包括容器的名称、镜像、环境变量和资源配置。

## 2. 卷

- [Volumes](Volumes.md) 介绍 `spec.volumes`，以及容器如何通过 `volumeMounts` 按名称挂载卷。

## 3. 进程命名空间

- [ShareProcessNamespace](ShareProcessNamespace.md) 介绍如何让同一个 Pod 内的容器共享进程命名空间。
