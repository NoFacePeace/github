# 架构

架构（Architecture）描述系统的整体结构，包括组件的职责划分、组件之间的关系与交互方式，以及指导系统设计和演进的原则。这里涵盖软件架构、系统架构等内容。

- [应用](Application.md)：应用的职责分配与协作方式，包括 C/S 和 B/S 架构。
- [存储](storage/Storage.md)：数据持久化、日志与故障恢复等通用设计机制。
- [检索](retrieval/Retrieval.md)：索引构建、查询处理、候选召回与结果排序等组件的职责和协作方式。
- [故障检测](failure-detection/FailureDetection.md)：通过心跳、健康探测等信号判断进程、节点或服务是否可能异常。
- [协调](coordination/Coordination.md)：管理共享资源与执行权的归属、交接和失效隔离。

## 1. 参考资料

- [MicrosoftDocs：Azure Architecture Center](https://github.com/MicrosoftDocs/architecture-center/tree/main)：微软 Azure 架构中心的文档源码仓库，涵盖架构风格、设计模式、参考架构、技术选型和最佳实践。
