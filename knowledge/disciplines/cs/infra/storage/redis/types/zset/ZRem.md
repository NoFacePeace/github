---
color: "#5B6CDA"
---

# ZREM

`ZREM` 从 ZSet 中删除一个或多个指定成员，同时移除它们关联的分数，返回实际删除的成员数量。

## 1. 命令格式

```text
ZREM key member [member ...]
```

| 参数 | 含义 |
| --- | --- |
| `key` | 有序集合的键名。 |
| `member` | 要删除的成员，至少提供一个，可以一次提供多个。 |

`ZREM` 不需要传入 score，也没有 `NX`、`XX` 等选项。key 后的所有参数都作为成员名解析。

## 2. 删除成员

以下示例假设 `ranking` 初始不存在：

```redis
ZADD ranking 10 alice 20 bob 30 carol
ZREM ranking alice bob
ZRANGE ranking 0 -1 WITHSCORES
```

`ZREM` 返回 `2`，表示删除了两个成员。集合中只剩 `carol`，分数为 `30`。

它可以删除任意指定成员，不要求该成员位于排序的头部或尾部；其余成员保持各自的分数和相对排序。

## 3. 返回值与边界

| 情况 | 结果 |
| --- | --- |
| 成员存在 | 删除该成员，计入返回数量。 |
| 成员不存在 | 忽略，不计入返回数量。 |
| key 不存在 | 返回 `0`，不创建集合。 |
| key 已存在但不是 ZSet | 返回 `WRONGTYPE` 错误。 |
| 删除最后一个成员 | Redis 自动删除这个空集合的 key。 |

```redis
ZREM ranking carol missing
EXISTS ranking
ZREM ranking carol
```

接上节示例，第一条返回 `1`；`ranking` 已为空并被删除，因此 `EXISTS` 返回 `0`；再次删除返回 `0`。

重复传入同一个成员也只会实际删除一次，因此返回值不一定等于传入的成员参数数量。`ZREM` 返回的是数量，不返回被删除成员及其分数。

## 4. 用作队列出队

```redis
ZADD pending 1000 session-a 2000 session-b 3000 session-c
ZREM pending session-b
ZRANGE pending 0 -1
```

删除后剩下 `session-a`、`session-c`。这是一种按 ID 移除待办的方式，不是只能移除队首的弹出操作。

典型处理流程为：

```text
读取待办 → 获取处理权 → 处理成功 → ZREM 移除
```

`ZREM` 本身只移除成员，不会检查处理权、释放锁、删除关联详情 KV 或停止执行进程。需要这些行为时，应由业务代码明确执行；如需在校验处理权后安全移除，可使用 Lua 脚本组合检查和删除。

## 5. 与其他删除方式的区别

| 命令 | 删除依据 | 返回值 |
| --- | --- | --- |
| `ZREM` | 指定成员名。 | 删除数量。 |
| `ZREMRANGEBYSCORE` | 指定分数范围。 | 删除数量。 |
| `ZREMRANGEBYRANK` | 指定排名范围。 | 删除数量。 |
| `ZPOPMIN`、`ZPOPMAX` | 最小或最大分数的成员。 | 弹出的成员及分数。 |
| `DEL` | 整个 Redis key。 | 实际删除的 key 数量。 |

若需要“取出成员并删除”，不要简单地在客户端先查询再执行 `ZREM` 并假设没有竞争；可考虑弹出命令或 Lua 脚本，并按业务需要补充确认与失败恢复机制。

## 6. 时间复杂度

常规复杂度为 `O(M log N)`，其中 `N` 为集合中的成员总数，`M` 为要删除的成员数量；删除单个成员通常记为 `O(log N)`。

实际成本还受底层编码影响，listpack 编码下可能涉及成员查找和连续内存的数据移动。
