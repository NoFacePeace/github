# 远程沙箱实现

远程沙箱方案需要区分完整平台、集群编排和底层隔离技术。它们解决的问题不同，可以组合使用；提供隔离能力不代表已经提供故障接管能力。

## 1. 平台服务

以下为常见方案示例，不代表市场份额排名。公开能力核对日期：2026-11-05；版本、区域和套餐差异需要按部署环境再次确认。

| 方案 | 主要能力 | 适用方向 |
|---|---|---|
| E2B | 通过 Python、JS/TS SDK 创建 Linux 沙箱，执行命令、操作文件，支持代码解释器及暂停恢复 | Agent 代码执行、数据分析和工具运行 |
| Daytona | 管理沙箱环境，提供命令、文件、Git、终端及预览等接口；能力随版本和沙箱类型变化 | 编程 Agent、构建测试和开发环境 |
| Modal Sandboxes | 在 Modal 计算平台运行不可信代码，结合镜像、持久化卷及可选 GPU 资源 | 弹性计算、GPU 任务及已有 Modal 工作流 |

E2B 文档说明暂停可以保存文件系统与内存。Daytona 的容器、VM 和 GPU 沙箱具有不同的持久化语义。Modal 的快照、Volumes 和函数重试属于不同能力，不能互相替代。

## 2. 底层隔离技术

| 技术 | 隔离方式 | 集成与取舍 |
|---|---|---|
| Docker / containerd 常规 Linux 容器 | 基于 namespace、cgroup、权限及系统调用限制等机制，通常共享宿主内核 | 生态成熟、开销低；需要配置挂载、网络、权限和配额，不能默认等同于 VM 隔离 |
| gVisor | 用户态应用内核承接应用的系统接口，减少工作负载直接接触宿主内核的范围 | 通过 `runsc` 接入 Docker、Kubernetes；需要验证兼容性与性能 |
| Kata Containers | 在轻量虚拟机内运行容器，客户机具有独立内核 | 可接入容器与 Kubernetes 体系，需要硬件虚拟化支持 |
| Firecracker | 基于 KVM 的精简虚拟机监控器，创建 microVM | 适合构建高密度隔离平台，但需配套镜像、网络、存储、调度和生命周期管理 |

Firecracker 是虚拟机底座；Kata 是容器运行时集成方案，可以选择 Firecracker 等虚拟机监控器。它们不是同一层级的替代品。

## 3. 集群编排与选型

已有 Kubernetes 集群时，可使用 [Agent Sandbox](../../../../../infra/kubernetes/ecosystem/agent-sandbox/AgentSandbox.md) 管理执行环境，再通过 RuntimeClass 选择相应隔离运行时。控制器负责环境生命周期，运行时负责隔离，存储系统负责数据持久化。

| 需求 | 可优先评估的方向 |
|---|---|
| 快速接入远程代码执行 | E2B、Daytona 等平台 SDK |
| GPU 或已有弹性计算工作流 | Modal 等计算平台的沙箱能力 |
| 在已有 K8s 内部自建 | 沙箱控制器配合 gVisor 或 Kata |
| 自建多租户执行平台 | Firecracker 等隔离底座，配套完整控制面 |

选型还应核对网络出口策略、凭据注入、文件持久化、并发与配额、空闲回收、检查点、节点故障恢复，以及实际恢复时间。仅比较启动速度或是否支持快照，不足以判断完整平台能力。

## 4. 参考资料

- [E2B 文档](https://docs.e2b.dev/)
- [E2B 暂停与恢复](https://docs.e2b.dev/sandbox/persistence)
- [Daytona 持久化与沙箱类型](https://www.daytona.io/docs/en/persistence/)
- [Modal Sandboxes](https://modal.com/docs/guide/sandboxes)
- [gVisor 原理与集成](https://gvisor.dev/docs/)
- [Kata Containers 架构与要求](https://katacontainers.io/software/)
- [Firecracker](https://firecracker-microvm.github.io/)
