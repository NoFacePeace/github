---
color: "#5B6CDA"
---

# Redis

Redis 是以内存为主要数据存储介质的键值存储系统，支持多种类型、持久化、复制和集群，常用于缓存、计数器、排行榜和消息流等场景。

## 1. 类型

- [类型](types/Types.md)：String、List、Hash、Set、ZSet 和 Stream 的语义、命令及底层编码。

## 2. 结构

- [结构](structs/Structs.md)：SDS、字典、整数集合、listpack、quicklist、跳表和基数树的作用与实现。

## 3. 可靠性

- [可靠性](reliability/Reliability.md)：数据持久化与故障后的恢复。

## 4. 配置

- [配置](config/Config.md)：配置文件、运行时配置，以及 AOF 开关与同步策略。
