---
color: "#5B6CDA"
---

# Redis

Redis 是以内存为主要数据存储介质的键值存储系统，支持多种类型、持久化、复制和集群，常用于缓存、计数器、排行榜和消息流等场景。

Redis 的数据在顶层以键值（Key-Value）形式组织：键是二进制安全的字符串，值可以是 String、List、Hash、Set、ZSet 或 Stream 等类型。KV 描述顶层组织方式，类型描述值的结构与支持的操作；例如 `user:1` 可以对应一个 Hash，其内部再以字段与值的映射保存用户信息。Pub/Sub 则是通信机制，不属于键的数据类型。

## 1. 类型

- [类型](types/Types.md)：String、List、Hash、Set、ZSet 和 Stream 的语义、命令及底层编码。

## 2. 结构

- [结构](structs/Structs.md)：SDS、字典、整数集合、listpack、quicklist、跳表和基数树的作用与实现。

## 3. 可靠性

- [可靠性](reliability/Reliability.md)：数据持久化与故障后的恢复。

## 4. 配置

- [配置](config/Config.md)：配置文件、运行时配置、AOF 和内存相关配置项。

## 5. 事务

- [事务](transactions/Transactions.md)：命令排队与连续执行、WATCH 乐观锁，以及取消执行与回滚的区别。

## 6. Lua

- [Lua](lua/Lua.md)：服务端脚本、条件判断与数据修改，以及错误处理和回滚限制。

## 7. 场景

- [场景](scenarios/Scenarios.md)：缓存、计数、限流、消息流和分布式锁等常见用途。

## 8. 集群

- [集群](cluster/Cluster.md)：数据分片、哈希槽、客户端路由和故障切换。

## 9. 内存

- [内存](memory/Memory.md)：内存上限、淘汰策略及容量观察。
