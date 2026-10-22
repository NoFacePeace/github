---
color: "#5B6CDA"
---

# Commands

ZSet 命令使用 `Z` 前缀。每个成员唯一，并关联一个分数；默认按分数升序排列，分数相同时按成员字节的字典序排列。

## 1. 命令概览

| 命令 | 作用 |
| --- | --- |
| [ZADD](ZAdd.md) | 添加成员，或更新已有成员的分数。 |
| [ZREM](ZRem.md) | 删除指定成员。 |
| `ZSCORE`、`ZMSCORE` | 查询一个或多个成员的分数。 |
| `ZINCRBY` | 增加或减少成员分数。 |
| [ZRANGE](ZRange.md) | 按排名、分数或字典序查询范围。 |
| `ZRANK`、`ZREVRANK` | 查询成员的升序或降序排名。 |
| `ZCARD` | 查询成员总数。 |
| `ZCOUNT` | 统计分数区间内的成员数。 |
| `ZREMRANGEBYSCORE`、`ZREMRANGEBYRANK` | 按分数或排名范围删除成员。 |
| `ZPOPMIN`、`ZPOPMAX` | 取出并删除最小或最大分数的成员。 |
| `BZPOPMIN`、`BZPOPMAX` | 没有成员时阻塞等待的弹出命令。 |
| `ZSCAN` | 使用游标增量遍历成员及分数。 |

## 2. 添加与更新

```redis
ZADD leaderboard 10 alice 20 bob 30 carol
ZADD leaderboard 25 alice
```

第一条命令添加三个成员；第二条更新 `alice` 的分数，不会新增第二个 `alice`。`ZADD` 默认返回新增成员数量，因此仅更新已有成员时返回 `0`。

常用选项：

| 选项 | 含义 |
| --- | --- |
| `NX` | 仅添加不存在的成员。 |
| `XX` | 仅更新已有成员。 |
| `GT` | 对已有成员，仅在新分数更大时更新；不阻止添加新成员。 |
| `LT` | 对已有成员，仅在新分数更小时更新；不阻止添加新成员。 |
| `CH` | 返回新增及分数发生变化的成员总数。 |

```redis
ZADD leaderboard NX 15 dave
ZADD leaderboard XX GT CH 40 alice
```

`NX` 与 `XX` 互斥；`GT`、`LT`、`NX` 也不能组合使用。若要限制为已有成员，可像上例一样将 `XX` 与 `GT` 组合。`GT`、`LT` 从 Redis 6.2 起提供。

## 3. 分数查询与增减

```redis
ZSCORE leaderboard alice
ZMSCORE leaderboard alice bob missing
ZINCRBY leaderboard 5 alice
ZINCRBY leaderboard -2 bob
```

`ZSCORE` 返回成员分数，不存在时返回空响应；`ZMSCORE` 按传入成员顺序返回对应分数，不存在的成员对应空响应。`ZMSCORE` 从 Redis 6.2 起提供。

`ZINCRBY` 返回更新后的分数。成员不存在时，会以增量为初始分数创建成员。

## 4. 范围查询

### 4.1 按排名查询

```redis
ZRANGE leaderboard 0 -1 WITHSCORES
ZRANGE leaderboard 0 2 REV WITHSCORES
```

- `0` 是第一个位置，排名从 `0` 开始。
- `-1` 是最后一个位置；起止位置均包含在结果中。
- `WITHSCORES` 同时返回分数，客户端可能将响应转换为不同的数据结构。
- `REV` 使用降序；分数相同时，成员字典序也反转。

第一条读取全部成员；第二条读取降序前三名。全量读取会随集合规模增大而增加开销。

### 4.2 按分数查询

```redis
ZRANGE leaderboard 10 30 BYSCORE WITHSCORES
ZRANGE leaderboard 30 10 BYSCORE REV WITHSCORES
ZRANGE leaderboard -inf +inf BYSCORE LIMIT 0 10
ZRANGE leaderboard (10 30 BYSCORE
```

分数边界默认包含端点，`(10` 表示不包含 `10`。`-inf`、`+inf` 表示负无穷与正无穷。命令行 shell 可能将括号视为特殊字符，通过 `redis-cli` 执行时应给该参数加引号，例如 `'(10'`。

`REV` 模式下，先给较大边界，再给较小边界。`LIMIT offset count` 用于限制返回范围；很大的 `offset` 仍可能产生较高的遍历开销。

`ZRANGE` 的 `BYSCORE`、`BYLEX`、`REV`、`LIMIT` 选项从 Redis 6.2 起提供。旧版本可使用 `ZRANGEBYSCORE`、`ZREVRANGE` 等命令。

### 4.3 按字典序查询

```redis
ZADD names 0 alice 0 bob 0 carol
ZRANGE names [alice [carol BYLEX
```

按字典序查询应在所有成员分数相同的集合中使用；分数不同时不能依赖其结果。`[` 表示包含边界，`(` 表示排除边界，`-` 与 `+` 分别表示无界的最小值和最大值。

## 5. 排名与数量

```redis
ZRANK leaderboard alice
ZREVRANK leaderboard alice
ZCARD leaderboard
ZCOUNT leaderboard 10 30
```

排名从 `0` 开始，成员不存在时返回空响应。`ZCARD` 返回整个集合的成员数量，`ZCOUNT` 返回指定分数区间内的数量。

## 6. 删除与弹出

```redis
ZREM leaderboard alice bob
ZREMRANGEBYSCORE leaderboard -inf 10
ZREMRANGEBYRANK leaderboard 0 2
ZPOPMIN leaderboard 1
ZPOPMAX leaderboard 1
```

前三种删除命令返回实际删除数量。弹出命令则在删除的同时返回成员与分数；`ZRANGE` 只读取，不会删除。

```redis
BZPOPMIN pending 5
```

该命令尝试取出 `pending` 中分数最小的一个成员，没有成员时最多等待 5 秒；超时返回空响应，超时参数为 `0` 时无限等待。

弹出并不等于可靠消费：客户端在取出后崩溃，成员不会自动重新入队。需要故障接管时，应额外设计处理中记录、确认和重试，或采用支持待确认记录的消息结构。

## 7. 增量遍历

```redis
ZSCAN leaderboard 0 COUNT 100
```

返回下一次使用的游标与一批成员及分数。继续传入返回的游标，直到游标回到 `0` 才表示一轮遍历结束。

`COUNT` 是工作量提示，不是返回数量的严格上限；单次可能返回空批次，紧凑编码下也可能一次返回全部成员。遍历可能重复返回成员，且不是一致性快照，集合变化时需考虑去重与并发修改。

## 8. 用作待调度队列

以入队时间为分数、Session ID 为成员，可以实现有序的等待名单：

```redis
ZADD pending 1792630800000 session-a
ZADD pending 1792630801000 session-b
ZRANGE pending 0 -1
ZREM pending session-a
```

流程是先读取，再获取处理权并处理，成功后用 `ZREM` 移除。多个消费者可能读到相同成员，需要额外的锁、租约或原子领取机制协调处理。

分数相同时按成员字典序排序，因此使用毫秒时间戳不能保证同毫秒内严格按提交先后排序。对排序有更严格要求时，需要设计可区分顺序的分数或成员。
