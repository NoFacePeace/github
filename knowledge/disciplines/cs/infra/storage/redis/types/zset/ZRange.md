---
color: "#5B6CDA"
---

# ZRANGE

`ZRANGE` 读取 ZSet 中指定范围的成员，不会删除它们。默认按排名查询，也可以按分数或成员字典序查询。

## 1. 命令格式

```text
ZRANGE key start stop [BYSCORE | BYLEX] [REV] [LIMIT offset count] [WITHSCORES]
```

| 参数或选项 | 含义 |
| --- | --- |
| `key` | 有序集合的键名。 |
| `start`、`stop` | 查询边界，默认表示排名；使用 `BYSCORE` 时表示分数，使用 `BYLEX` 时表示字典序边界。 |
| `BYSCORE` | 按分数区间查询。 |
| `BYLEX` | 按成员字节字典序区间查询，应在所有成员分数相同的集合中使用。 |
| `REV` | 反转排序方向。 |
| `LIMIT offset count` | 跳过指定数量后，最多返回指定数量；只能用于 `BYSCORE` 或 `BYLEX` 查询。 |
| `WITHSCORES` | 在返回成员时同时返回分数。 |

`BYSCORE` 与 `BYLEX` 互斥。`BYSCORE`、`BYLEX`、`REV`、`LIMIT` 从 Redis 6.2 起提供；旧版本使用 `ZRANGEBYSCORE`、`ZRANGEBYLEX`、`ZREVRANGE` 等对应命令。

## 2. 按排名查询

以下示例假设 `ranking` 初始不存在：

```redis
ZADD ranking 10 alice 20 bob 30 carol 40 dave
ZRANGE ranking 0 1
ZRANGE ranking 0 -1
ZRANGE ranking -2 -1
```

结果分别为：

```text
alice、bob
alice、bob、carol、dave
carol、dave
```

排名从 `0` 开始，起止位置都包含在结果中。负数从尾部计数，`-1` 表示最后一个成员，`-2` 表示倒数第二个成员。

排名范围超出集合边界时会按有效范围处理；起点超过末尾或有效起点大于终点时返回空结果。key 不存在时返回空结果，key 类型不是 ZSet 时返回 `WRONGTYPE` 错误。

## 3. 返回分数与倒序

```redis
ZRANGE ranking 0 1 WITHSCORES
ZRANGE ranking 0 1 REV WITHSCORES
```

第一条返回最低的两个成员及分数，第二条返回最高的两个：

```text
升序：alice → 10，bob → 20
降序：dave → 40，carol → 30
```

`REV` 下，排名 `0` 表示倒序后的第一个成员，分数相同时成员字典序也反转。

在 RESP2 下，`WITHSCORES` 通常返回成员与分数交替排列的数组；RESP3 和客户端封装可能呈现为嵌套数组或其他结构，应以使用的协议及客户端为准。

## 4. 按分数查询

```redis
ZRANGE ranking 20 30 BYSCORE WITHSCORES
ZRANGE ranking (20 40 BYSCORE
ZRANGE ranking -inf +inf BYSCORE
```

第一条包含分数 `20` 到 `30`，返回 `bob`、`carol`。第二条排除 `20`、包含 `40`，返回 `carol`、`dave`。第三条覆盖全部分数。

- 默认包含端点；前缀 `(` 表示排除端点。
- `-inf`、`+inf` 可用作无穷边界。
- 倒序分数查询必须先写较大边界，再写较小边界：

```redis
ZRANGE ranking 40 20 BYSCORE REV
```

通过 shell 调用 `redis-cli` 时，需要给含括号的边界加引号，避免被 shell 解析：

```sh
redis-cli ZRANGE ranking '(20' 40 BYSCORE
```

## 5. LIMIT 分页

```redis
ZRANGE ranking -inf +inf BYSCORE LIMIT 1 2
```

表示在分数查询结果中跳过一个成员，再最多返回两个成员，接上例得到 `bob`、`carol`。`count` 为负数时返回跳过 offset 后的所有成员，应避免因此意外读取大量数据。

`LIMIT` 不能用于默认的排名查询；按排名分页时直接计算起止排名：

```redis
ZRANGE ranking 1 2
```

分页不是一致性快照。两次查询之间成员增删或分数变化，可能造成重复或遗漏；使用较大的 `offset` 也会增加跳过成员的遍历开销。

## 6. 按字典序查询

以下示例假设 `names` 初始不存在，所有成员分数均为 `0`：

```redis
ZADD names 0 alice 0 bob 0 carol 0 dave
ZRANGE names [alice [carol BYLEX
ZRANGE names (alice [carol BYLEX
ZRANGE names - + BYLEX
```

结果分别为 `alice、bob、carol`、`bob、carol` 和全部成员。

字典序边界中，`[` 表示包含端点，`(` 表示排除端点，`-` 和 `+` 表示无界的最小值与最大值。通过 shell 执行时，应给边界参数加引号。

字典序基于成员字节，不是语言或地区的自然排序。集合成员分数不同时，不应依赖 `BYLEX` 的结果。

## 7. 用于待调度队列

```redis
ZADD pending 1000 session-a 2000 session-b 3000 session-c
ZRANGE pending 0 -1
```

返回按入队时间排序的全部 Session ID。应用代码可以遍历结果，逐个尝试获取处理权；抢不到某项时继续下一项。

`ZRANGE` 只读取，不会占用任务或将其出队，所以多个 Server 可能读到相同成员。可靠处理需要额外的处理权、确认及故障恢复机制，处理完成后再移除成员。

## 8. 时间复杂度

常规复杂度为 `O(log N + M)`，`N` 为集合成员总数，`M` 为返回成员数量。按分数或字典序查询并使用较大的 `LIMIT offset` 时，还需考虑跳过成员的成本。

`ZRANGE key 0 -1` 会返回全部成员；集合较大时，会增加 Redis 执行、响应传输和客户端内存开销。应按需求限制读取范围，增量扫描场景可考虑游标遍历。
