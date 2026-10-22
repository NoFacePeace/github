---
color: "#5B6CDA"
---

# ZADD

`ZADD` 向 ZSet 添加成员及其分数，或更新已有成员的分数，并自动维护排序位置。

## 1. 命令格式

```text
ZADD key [NX | XX] [GT | LT] [CH] [INCR] score member [score member ...]
```

| 参数 | 含义 |
| --- | --- |
| `key` | 有序集合的键名。 |
| `score` | 成员的分数，支持负数、小数和正负无穷，不接受 NaN。 |
| `member` | 成员，在同一个 ZSet 中唯一。 |

注意参数顺序是先 `score`、后 `member`。键不存在时，允许新增的操作会创建 ZSet；键已存在但类型不是 ZSet 时返回类型错误。

## 2. 添加与更新

以下示例假设 `ranking` 初始不存在：

```redis
ZADD ranking 10 alice 20 bob
ZADD ranking 30 alice
ZSCORE ranking alice
ZCARD ranking
```

结果：

1. 第一条返回 `2`，新增两个成员。
2. 第二条返回 `0`，只更新 `alice` 的分数，没有新增成员。
3. `alice` 的分数变为 `30`。
4. 集合中仍只有两个成员。

默认情况下，返回值是新增成员数量，不包含已有成员的分数更新。返回 `0` 不一定表示没有修改数据。

## 3. 条件选项

| 选项 | 含义 |
| --- | --- |
| `NX` | 只添加不存在的成员，不修改已有成员。 |
| `XX` | 只修改已有成员，不添加新成员。 |
| `GT` | 对已有成员，只接受比当前分数更大的新分数；仍允许添加新成员。 |
| `LT` | 对已有成员，只接受比当前分数更小的新分数；仍允许添加新成员。 |

```redis
ZADD ranking NX 15 alice
ZADD ranking XX 25 bob
ZADD ranking XX GT 40 alice
ZADD ranking XX LT 10 bob
```

`XX GT` 表示仅允许已有成员提高分数，`XX LT` 表示仅允许已有成员降低分数。这里的提高、降低是数值比较，是否代表更高业务优先级需要另行约定。

选项约束：

- `NX` 和 `XX` 不能一起使用。
- `GT` 和 `LT` 不能一起使用。
- `NX` 不能与 `GT` 或 `LT` 一起使用。
- 分数相等时，`GT`、`LT` 都不会进行分数更新。
- `GT`、`LT` 从 Redis 6.2 起提供。

## 4. CH 返回变化数量

`CH` 表示返回新增和分数实际变化的成员总数，而不仅是新增数量。

以下示例假设 `scores` 初始不存在：

```redis
ZADD scores 10 alice
ZADD scores CH 20 alice 30 bob
ZADD scores CH 20 alice
```

第二条返回 `2`：`alice` 更新分数，`bob` 新增。第三条返回 `0`，因为 `alice` 的分数没有变化。

`CH` 只改变返回值的计数方式，不改变唯一性或排序规则。

## 5. INCR 按增量更新

```redis
ZADD ranking INCR 5 alice
ZADD ranking INCR -2 bob
```

`INCR` 将传入的 score 解释为增量，行为类似 `ZINCRBY`。成员不存在且选项允许新增时，以增量作为初始分数创建成员。

使用 `INCR` 时：

- 一次只能传入一组 score、member。
- 返回更新后的分数，不返回新增或变化数量，即使同时指定 `CH`。
- 条件不满足、操作未执行时返回空响应，例如对已有成员使用 `NX INCR`。
- `GT`、`LT` 比较增量计算后的新分数与原分数。

## 6. 用作队列入队

```redis
ZADD pending 1792630800000 session-a
ZADD pending 1792630801000 session-b
ZRANGE pending 0 -1
```

用入队时间作为 score、Session ID 作为 member，可以按时间读取待办。默认重复添加同一 member 会更新其时间并可能改变排序；如果需要保留第一次入队时间，可以使用 `NX`：

```redis
ZADD pending NX 1792630802000 session-a
```

分数相同的成员按字节字典序排列，不保证同一毫秒内按提交先后排序。`ZADD` 只负责修改集合，不负责消费者处理权、确认或失败重试。

## 7. 复杂度与执行边界

常规复杂度为每个添加或更新成员 `O(log N)`；一条命令处理 `M` 个成员时为 `O(M log N)`。listpack 编码下的具体成本还包含扫描和数据移动。

单条 `ZADD` 执行期间不会插入其他客户端的普通命令。与其他 Redis 命令组合使用时，如果先在客户端读取分数，再判断并执行 `ZADD`，其他客户端可能在两条命令之间修改数据；应优先考虑条件选项，复杂条件可通过 Lua 脚本将读取、判断与写入组合起来。
