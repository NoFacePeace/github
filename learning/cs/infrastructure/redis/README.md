# Redis 底层实现学习

记录 Redis 的架构设计、核心数据结构、命令执行流程与源码阅读笔记。

## 1. 源码基线

- 官方仓库：[redis/redis](https://github.com/redis/redis)
- 学习版本：[7.2.16](https://github.com/redis/redis/tree/7.2.16)
- 本地源码目录：`/Users/haotao.chen/Desktop/repositories/redis`

学习笔记保存在本目录，源码在独立仓库中阅读。

## 2. 学习顺序

1. **启动与命令执行**：服务初始化、客户端连接、命令解析与执行流程。
2. **数据结构与对象编码**：SDS、字典、整数集合、listpack、quicklist、跳表，以及各数据类型的编码选择与转换。
3. **事件循环与网络 I/O**：I/O 多路复用、文件事件、时间事件与 I/O 线程。
4. **内存管理**：内存分配、过期删除、内存淘汰与惰性释放。
5. **持久化**：RDB、AOF、重写流程与写时复制。
6. **复制与高可用**：全量同步、增量同步、复制积压缓冲区与 Sentinel 故障转移。
7. **集群**：哈希槽、请求路由、数据迁移与故障检测。

## 3. 数据类型

数据类型定义用户可以执行的操作，底层编码决定数据如何存储；同一种数据类型可以对应多种编码。

### 3.1 类型总览

| 数据类型 | 含义 | 常见用途 | 底层编码或数据结构 |
| --- | --- | --- | --- |
| String | 字符串，可存储数字或二进制数据 | 缓存、计数器 | `int` 整数编码，或基于 SDS 的 `embstr`、`raw` 编码 |
| List | 有序、允许重复的列表 | 队列、最近记录 | quicklist，节点通常使用 listpack 存储元素 |
| Hash | 字段与值的映射 | 用户信息、对象属性 | listpack 或哈希表 |
| Set | 无序、不重复的集合 | 去重、交集、并集、差集 | intset、listpack 或哈希表 |
| Sorted Set（ZSet） | 按分数排序、成员不重复的集合 | 排行榜、延迟任务 | listpack，或跳表与哈希表的组合 |
| Stream | 带消息 ID 的追加式消息流 | 消息消费、消费者组 | 基数树（rax）与 listpack 的组合 |

类型与编码定义见 [src/server.h](https://github.com/redis/redis/blob/7.2.16/src/server.h)。此外，`OBJ_MODULE` 用于模块自定义的数据类型。

### 3.2 数据类型与对象编码

Redis 使用 `redisObject` 表示对象：`type` 标识数据类型，`encoding` 标识底层编码，`ptr` 保存数据指针；整数编码会直接利用该字段保存整数值。

编码转换是同一个键在保持数据类型和内容不变的情况下，更换底层存储方式。

Redis 根据元素数量、元素大小和内容选择编码，在内存占用与操作效率之间取舍。例如，小 Hash 使用紧凑的 listpack，减少额外内存开销；超过相应配置阈值后转换为哈希表，提高字段查找效率。用户仍然使用 `HSET`、`HGET` 等命令，无须手动管理编码。

- **“listpack 或哈希表”**表示可选编码，一个键在某个时刻采用其中一种。
- **“跳表与哈希表的组合”**表示同时使用两种结构：ZSet 用哈希表按成员查找分数，用跳表维护排序和执行范围查询。
- 编码转换规则因类型而异，数据缩小后不一定自动转回紧凑编码。
- 可通过 `TYPE key` 查看类型，通过 `OBJECT ENCODING key` 查看当前编码。

### 3.3 String

- **基本语义**：存储字符串、数字或二进制数据。
- **底层实现**：学习 SDS，以及 `int`、`embstr`、`raw` 三种编码的内存布局。
- **编码转换**：关注整数编码条件、短字符串编码选择，以及追加或修改内容时的转换。
- **源码入口**：[t_string.c](https://github.com/redis/redis/blob/7.2.16/src/t_string.c)、[object.c](https://github.com/redis/redis/blob/7.2.16/src/object.c)、[sds.c](https://github.com/redis/redis/blob/7.2.16/src/sds.c)。

#### 3.3.1 常用命令

| 命令 | 用途 |
| --- | --- |
| `SET` / `GET` | 设置或读取值；SET 支持 NX、XX 条件及 EX、PX 过期选项。 |
| `MSET` / `MGET` | 批量设置或读取多个键。 |
| `INCR` / `DECR` / `INCRBY` / `DECRBY` | 对整数值递增、递减或按指定步长修改。 |
| `INCRBYFLOAT` | 对数值增加指定浮点增量。 |
| `APPEND` / `STRLEN` | 追加内容或获取字符串字节长度。 |
| `GETRANGE` / `SETRANGE` | 读取指定字节范围，或从指定偏移覆盖内容。 |
| `GETDEL` / `GETEX` | 读取并删除值，或读取时修改过期设置。 |

### 3.4 List

- **基本语义**：有序且允许重复，支持两端插入、弹出和范围读取。
- **底层实现**：学习 quicklist 如何组织节点，以及 listpack 如何紧凑存储元素。
- **存储变化**：关注节点容量、分裂与合并、压缩，以及大元素的独立存储。
- **源码入口**：[t_list.c](https://github.com/redis/redis/blob/7.2.16/src/t_list.c)、[quicklist.c](https://github.com/redis/redis/blob/7.2.16/src/quicklist.c)、[listpack.c](https://github.com/redis/redis/blob/7.2.16/src/listpack.c)。

#### 3.4.1 常用命令

| 命令 | 用途 |
| --- | --- |
| `LPUSH` / `RPUSH` | 向列表左端或右端插入元素。 |
| `LPOP` / `RPOP` | 从左端或右端移除并返回元素。 |
| `BLPOP` / `BRPOP` | 阻塞等待可弹出的元素。 |
| `LRANGE` / `LINDEX` | 按索引范围读取元素，或读取指定索引的元素。 |
| `LLEN` | 获取元素数量。 |
| `LSET` / `LREM` / `LTRIM` | 修改指定索引的元素、按值移除元素，或只保留指定索引范围。 |
| `LMOVE` / `BLMOVE` | 将源列表一端的元素原子移动到目标列表一端；BLMOVE 支持阻塞等待。 |

### 3.5 Hash

- **基本语义**：存储字段与值的映射。
- **底层实现**：学习 listpack 中字段与值的组织方式，以及哈希表的查找与扩容。
- **编码转换**：关注字段数量、字段长度和值长度如何触发 listpack 向哈希表转换。
- **源码入口**：[t_hash.c](https://github.com/redis/redis/blob/7.2.16/src/t_hash.c)、[dict.c](https://github.com/redis/redis/blob/7.2.16/src/dict.c)。

#### 3.5.1 常用命令

| 命令 | 用途 |
| --- | --- |
| `HSET` / `HGET` | 设置一个或多个字段，或读取一个字段。 |
| `HMGET` / `HGETALL` | 读取多个指定字段，或读取全部字段和值。 |
| `HDEL` / `HEXISTS` | 删除字段或判断字段是否存在。 |
| `HLEN` / `HKEYS` / `HVALS` | 获取字段数量、全部字段名或全部值。 |
| `HSETNX` | 仅在字段不存在时设置值。 |
| `HINCRBY` / `HINCRBYFLOAT` | 对字段值增加整数或浮点增量。 |
| `HSCAN` | 使用游标增量遍历字段和值。 |

### 3.6 Set

- **基本语义**：存储不重复的元素，不保证顺序。
- **底层实现**：学习 intset、listpack 和哈希表的存储方式，以及集合运算。
- **编码转换**：关注元素是否为整数、元素数量和大小对编码选择与转换的影响。
- **源码入口**：[t_set.c](https://github.com/redis/redis/blob/7.2.16/src/t_set.c)、[intset.c](https://github.com/redis/redis/blob/7.2.16/src/intset.c)。

#### 3.6.1 常用命令

| 命令 | 用途 |
| --- | --- |
| `SADD` / `SREM` | 添加或移除成员。 |
| `SISMEMBER` / `SMISMEMBER` | 判断一个或多个元素是否属于集合。 |
| `SCARD` / `SMEMBERS` | 获取成员数量或全部成员。 |
| `SINTER` / `SUNION` / `SDIFF` | 计算交集、并集或差集；差集以第一个集合为基准。 |
| `SINTERSTORE` / `SUNIONSTORE` / `SDIFFSTORE` | 将集合运算结果保存到目标键。 |
| `SPOP` / `SRANDMEMBER` | 随机移除并返回成员，或仅随机读取成员。 |
| `SMOVE` | 将成员从源集合原子移动到目标集合。 |
| `SSCAN` | 使用游标增量遍历成员。 |

### 3.7 ZSet

- **基本语义**：成员唯一，按分数排序。
- **底层实现**：学习 listpack 编码，以及跳表与哈希表如何协作支持成员查找和有序范围查询。
- **编码转换**：关注成员数量与长度如何触发 listpack 向跳表与哈希表组合的转换。
- **源码入口**：[t_zset.c](https://github.com/redis/redis/blob/7.2.16/src/t_zset.c)。

#### 3.7.1 常用命令

| 命令 | 用途 |
| --- | --- |
| `ZADD` / `ZREM` | 添加成员或更新分数，或移除成员。 |
| `ZSCORE` / `ZMSCORE` | 读取一个或多个成员的分数。 |
| `ZINCRBY` | 增加成员的分数。 |
| `ZRANGE` | 按排名读取成员，也支持 BYSCORE、BYLEX 范围及 REV 逆序选项；按字典序查询适用于成员分数相同的情况。 |
| `ZRANK` / `ZREVRANK` | 获取成员的正序或逆序排名，排名从 0 开始。 |
| `ZCARD` / `ZCOUNT` | 获取成员总数，或指定分数范围内的成员数量。 |
| `ZPOPMIN` / `ZPOPMAX` | 移除并返回最低分或最高分的成员。 |
| `BZPOPMIN` / `BZPOPMAX` | 阻塞等待并弹出最低分或最高分的成员。 |
| `ZREMRANGEBYRANK` / `ZREMRANGEBYSCORE` | 按排名范围或分数范围移除成员。 |
| `ZSCAN` | 使用游标增量遍历成员和分数。 |

### 3.8 Stream

- **基本语义**：按消息 ID 组织消息，支持消费者组。
- **底层实现**：学习基数树与 listpack 如何组织消息，以及消费者组、待确认消息列表（PEL）的管理。
- **存储变化**：关注节点增长、消息删除与裁剪，不套用其他类型的紧凑编码转换规则。
- **源码入口**：[t_stream.c](https://github.com/redis/redis/blob/7.2.16/src/t_stream.c)、[rax.c](https://github.com/redis/redis/blob/7.2.16/src/rax.c)。

#### 3.8.1 常用命令

| 命令 | 用途 |
| --- | --- |
| `XADD` | 追加消息，可使用 * 自动生成消息 ID。 |
| `XLEN` / `XRANGE` / `XREVRANGE` | 获取消息数量，或按消息 ID 范围正序、逆序读取。 |
| `XREAD` | 从指定消息 ID 之后读取消息，支持 BLOCK 阻塞等待。 |
| `XGROUP CREATE` | 创建消费者组，设置组的起始读取位置。 |
| `XREADGROUP` | 以消费者组身份读取消息；通常使用 > 获取尚未投递给组内消费者的新消息。 |
| `XACK` | 确认消息处理完成，从组的待确认消息列表中移除记录，不删除流中的消息。 |
| `XPENDING` | 查看消费者组的待确认消息。 |
| `XCLAIM` / `XAUTOCLAIM` | 接管满足空闲时间条件的待确认消息，用于处理消费者失联等情况。 |
| `XDEL` / `XTRIM` | 删除指定消息，或按长度、最小消息 ID 裁剪消息流。 |
| `XINFO` | 查看流、消费者组或消费者的信息。 |

### 3.9 扩展功能

| 功能 | 用途 | 实际基于的数据类型 |
| --- | --- | --- |
| Bitmap | 按位存储状态，如签到记录 | String |
| HyperLogLog | 近似统计不重复元素数量 | String |
| Geo | 地理位置存储与附近搜索 | ZSet |

这些功能拥有专门的操作命令，但不是独立的核心对象类型。相关实现见 [bitops.c](https://github.com/redis/redis/blob/7.2.16/src/bitops.c)、[hyperloglog.c](https://github.com/redis/redis/blob/7.2.16/src/hyperloglog.c) 和 [geo.c](https://github.com/redis/redis/blob/7.2.16/src/geo.c)。

## 4. 数据结构

本章从存储结构出发，学习内存布局、查找与更新过程，以及扩容和空间优化。同一种结构可以被多个数据类型复用。

### 4.1 结构总览

| 数据结构 | 核心特点 | 主要关联 |
| --- | --- | --- |
| SDS | 带长度和容量信息的动态字符串 | String，以及键名、字段名等字符串数据 |
| 字典（dict） | 基于哈希表的键值映射 | 数据库键空间、Hash、Set、ZSet |
| 整数集合（intset） | 有序、紧凑存储整数 | Set |
| listpack | 在连续内存中紧凑存储多个元素 | List、Hash、Set、ZSet、Stream |
| quicklist | 将多个节点组织成双向链表，节点通常持有 listpack | List |
| 跳表（skiplist） | 通过多层索引支持有序查找 | ZSet |
| 基数树（rax） | 压缩公共前缀的树形索引 | Stream 的消息块索引及内部管理结构 |

### 4.2 SDS：动态字符串

- **作用**：保存二进制安全的字符串，显式记录长度，避免获取长度时扫描整个字符串。
- **学习重点**：不同长度的头部布局、已用长度与可用容量、扩容策略，以及与 C 字符串的关系。
- **关联类型**：String 的 `raw`、`embstr` 编码使用 SDS；两种编码的对象与字符串内存分配方式不同。
- **源码入口**：[sds.h](https://github.com/redis/redis/blob/7.2.16/src/sds.h)、[sds.c](https://github.com/redis/redis/blob/7.2.16/src/sds.c)。

### 4.3 字典：哈希表

- **作用**：通过哈希函数定位桶，支持平均 O(1) 的查找、插入和删除。
- **学习重点**：桶与条目的组织、哈希冲突处理、负载因子、扩缩容，以及渐进式 rehash 如何分摊迁移成本。
- **关联类型**：Hash 存储字段和值，Set 存储成员，ZSet 借助字典按成员查分数；数据库本身也使用字典管理键。
- **源码入口**：[dict.h](https://github.com/redis/redis/blob/7.2.16/src/dict.h)、[dict.c](https://github.com/redis/redis/blob/7.2.16/src/dict.c)。

### 4.4 intset：整数集合

- **作用**：在连续内存中按顺序存储不重复的整数，减少通用哈希表的额外开销。
- **学习重点**：二分查找、插入与删除时的数据移动，以及 16、32、64 位整数存储宽度的升级。
- **关联类型**：Set 的一种底层编码。intset 内部的整数宽度升级与 Set 转为其他编码是两个不同过程。
- **源码入口**：[intset.h](https://github.com/redis/redis/blob/7.2.16/src/intset.h)、[intset.c](https://github.com/redis/redis/blob/7.2.16/src/intset.c)。

### 4.5 listpack：紧凑列表

- **作用**：把多个整数或字符串元素存入一段连续内存，减少逐个分配对象和保存指针的开销。
- **学习重点**：头部、元素编码、反向遍历所需的长度信息，以及插入、删除时的内存移动。
- **关联类型**：Hash、Set、ZSet 可直接使用 listpack；quicklist 和 Stream 则将它作为内部存储单元。
- **源码入口**：[listpack.h](https://github.com/redis/redis/blob/7.2.16/src/listpack.h)、[listpack.c](https://github.com/redis/redis/blob/7.2.16/src/listpack.c)。

### 4.6 quicklist：分块双向链表

- **作用**：将数据分散到多个双向链接的节点，在紧凑存储和局部修改成本之间取舍；普通节点持有 listpack，大元素也可独立存储。
- **学习重点**：节点布局、两端操作、节点分裂与合并、容量限制，以及中间节点的压缩策略。
- **关联类型**：List。quicklist 是组织节点的外层结构，listpack 是节点内的存储结构。
- **源码入口**：[quicklist.h](https://github.com/redis/redis/blob/7.2.16/src/quicklist.h)、[quicklist.c](https://github.com/redis/redis/blob/7.2.16/src/quicklist.c)。

### 4.7 skiplist：跳表

- **作用**：用多层前向指针加速有序查找，查找、插入和删除的期望复杂度为 O(log N)。
- **学习重点**：随机层高、按分数及成员排序、跨度（span）如何支持排名计算，以及范围遍历。
- **关联类型**：ZSet 的非 listpack 编码同时使用跳表和字典，分别支持有序访问与按成员查找。
- **源码入口**：[t_zset.c](https://github.com/redis/redis/blob/7.2.16/src/t_zset.c)，重点关注 `zslCreate`、`zslInsert`、`zslDelete` 和 `zslGetRank`。

### 4.8 rax：基数树

- **作用**：按键的字节序列建立索引，压缩公共前缀和单一路径，支持有序遍历。
- **学习重点**：压缩节点布局、路径匹配、插入时的节点分裂、删除与迭代器。
- **关联类型**：Stream 使用 rax 索引消息块，消息块内部通过 listpack 保存消息；消费者组的部分管理结构也使用 rax。
- **源码入口**：[rax.h](https://github.com/redis/redis/blob/7.2.16/src/rax.h)、[rax.c](https://github.com/redis/redis/blob/7.2.16/src/rax.c)。

## 5. 缓存应用与常见问题

本章结合 Redis 学习缓存设计，关注请求如何在应用、缓存和数据库之间流转，以及失效、并发和故障情况下的处理方式。

### 5.1 缓存穿透

- **问题**：请求查询缓存和数据库中都不存在的数据，反复绕过缓存访问数据库。
- **学习重点**：参数校验、空值缓存、布隆过滤器，以及误判、数据新增后过滤器更新和空值过期策略。

### 5.2 缓存击穿

- **问题**：热点键失效时，大量并发请求同时查询数据库并重建缓存。
- **学习重点**：互斥重建、请求合并、逻辑过期与后台刷新，以及等待超时、重建失败和旧数据可接受程度。

### 5.3 缓存雪崩

- **问题**：大量缓存集中失效，或缓存服务不可用，导致请求集中访问数据库。
- **学习重点**：过期时间打散、缓存预热、多级缓存、限流降级与故障恢复；区分批量过期和服务故障两种场景。

### 5.4 缓存与数据库一致性

- **问题**：缓存和数据库是两份数据，更新顺序、并发读写和操作失败都可能导致缓存保留旧值。
- **学习重点**：Cache Aside（旁路缓存）、先更新数据库再删除缓存、删除失败重试，以及通过变更订阅更新或失效缓存。
- **分析方式**：画出并发时序，明确每一步的失败情况与不一致窗口；不要把延迟双删或某一种更新顺序视为强一致保证。

### 5.5 热点 Key

- **问题**：访问流量集中在少数键上，使单个节点或分片承受过高负载；即使键没有过期也可能发生。
- **学习重点**：热点识别、本地缓存、请求合并、读流量分散，以及副本或多份缓存引入的一致性与失效管理。

### 5.6 大 Key

- **问题**：单个键的值过大或集合元素过多，增加内存、网络传输、命令执行和删除的成本。
- **学习重点**：结合字节大小、元素数量和实际操作识别大键，学习数据拆分、分批读取、渐进遍历和 `UNLINK` 异步释放。
- **分析方式**：区分“大键”和“慢操作”，同时检查命令复杂度、返回数据量与访问频率。

## 6. 笔记约定

- 以具体问题为入口，结合数据结构和调用链解释实现。
- 引用源码时注明文件和函数，并附上源码链接。
- 区分源码事实、个人理解与实验结论。
- 涉及其他版本时，明确记录与学习基线的差异。
