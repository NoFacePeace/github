---
color: "#5B6CDA"
---

# Kubernetes

Kubernetes（K8s）是容器编排平台，用于管理容器部署、扩缩容和故障恢复。它通常通过 containerd 或 CRI-O 等容器运行时运行容器。

## 1. 与 Docker 的关系

Kubernetes 可以运行 Docker 构建的兼容镜像，但不要求安装 Docker。两者分别承担容器编排与镜像构建、容器管理等职责。

- [Docker](../docker/Docker.md)

## 2. 环境变量

- [环境变量](env/Env.md)：Kubernetes 客户端常用环境变量。

## 3. 生态

- [Kubernetes 生态](ecosystem/Ecosystem.md)：相关项目、标准与接口、托管服务与发行版、社区组织，以及集成方案和选型对比。

## 4. Service

- [Service](service/Service.md)：Pod 的稳定访问入口、后端选择与服务发现。

## 5. 资源与对象

- [资源与对象](resource/Resource.md)：Kubernetes API 对象的结构、元数据、标签和注解。

## 6. Pod

- [Pod](pod/Pod.md)：Kubernetes 的部署单元、容器配置和 Downward API。

## 7. 持久存储

- [持久存储](storage/Storage.md)：PV、PVC、Pod 挂载与动态供给。

## 8. 准入控制

- [准入控制](admission/Admission.md)：API Server 在保存资源前如何变更和验证请求对象。
