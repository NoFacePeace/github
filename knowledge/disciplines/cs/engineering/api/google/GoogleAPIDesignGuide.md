# Google API 设计指南

Google API Design Guide 是 Google 面向网络 API 的通用设计指南，用于指导 Cloud API 和其他 Google API 的设计。它适用于 REST API 和 RPC API，尤其关注 gRPC API。

## 1. 官方入口

- [API Design Guide](https://docs.cloud.google.com/apis/design)：设计指南入口，按主题链接到相关设计规则。
- [API Improvement Proposals（AIP）](https://google.aip.dev/)：记录 Google API 设计决策和具体设计指导的文档集合。

## 2. 指南与 AIP 的关系

API Design Guide 提供整体入口和主题导航，许多具体规则通过编号的 AIP 文档展开，例如资源命名、标准方法、错误处理和向后兼容。

AIP 不仅是设计指导，也是管理和追踪这些指导的机制。其他组织可以参考并选择适合自身 API 的规则，不必将全部条目直接套用到每个项目。

## 3. 常用主题

| 主题 | 官方条目 |
| --- | --- |
| 面向资源的设计 | [AIP-121](https://google.aip.dev/121) |
| 资源命名 | [AIP-122](https://google.aip.dev/122) |
| 方法概述 | [AIP-130](https://google.aip.dev/130) |
| 标准方法 | [Get](https://google.aip.dev/131)、[List](https://google.aip.dev/132)、[Create](https://google.aip.dev/133)、[Update](https://google.aip.dev/134)、[Delete](https://google.aip.dev/135) |
| 自定义方法 | [AIP-136](https://google.aip.dev/136) |
| 向后兼容 | [AIP-180](https://google.aip.dev/180) |
| 版本管理 | [AIP-185](https://google.aip.dev/185) |
| 命名约定 | [AIP-190](https://google.aip.dev/190) |
| 错误处理 | [AIP-193](https://google.aip.dev/193) |
