---
color: "#5B6CDA"
---

# Model

ZSet 的数据模型分为两层：外层是 Redis 的键值关系，键对应整个有序集合；集合内部由唯一成员（member）及其分数（score）组成。

## 1. 整体结构

```text
Redis Key：leaderboard
    ↓
Value：一个 ZSet
    ├─ alice → score 10
    ├─ bob   → score 20
    └─ carol → score 20
```

| 概念 | 含义 | 示例 |
| --- | --- | --- |
| key | 整个 ZSet 在 Redis 中的键名。 | `leaderboard`。 |
| [member](Member.md) | 集合中的成员，在同一个 ZSet 内唯一。 | `alice`、`bob`。 |
| [score](Score.md) | 与成员关联的数值，决定排序位置，可以重复。 | `10`、`20`。 |

可以将内部关系类比为 `member → score`，但它不是可以保存任意值的通用 KV：member 是二进制安全的字符串，score 是双精度浮点数。score 不接受 NaN，可使用正负无穷；需要精确区分整数时，应注意双精度表示的精度范围。

## 2. 成员唯一性

同一个 ZSet 内，member 不能重复。默认情况下，再次添加相同 member 会更新它的 score：

```redis
ZADD leaderboard 10 alice
ZADD leaderboard 20 alice
ZCARD leaderboard
ZSCORE leaderboard alice
```

最终成员数量为 `1`，`alice` 的分数为 `20`，不会保留两个 `alice`。不同 ZSet 可以包含相同的 member。

成员是否相同取决于完整字节内容，例如 `alice` 和 `Alice` 是不同成员。

## 3. 分数与排序

不同成员可以拥有相同 score。默认顺序为：

1. 按 score 从小到大排序。
2. score 相同时，按 member 的字节字典序排序。

```redis
ZADD ranking 20 bob 10 carol 20 alice
ZRANGE ranking 0 -1 WITHSCORES
```

逻辑结果为：

```text
carol → 10
alice → 20
bob   → 20
```

成员的插入先后不决定同分顺序。因此，用入队时间戳作为 score 时，同一毫秒入队的成员不一定按提交先后排列。

## 4. 成员与业务数据

ZSet 没有独立于 member、score 之外的 payload 字段。实际业务通常用 member 保存 ID，再用其他 Redis 键或数据库记录保存详情。

```text
ZSet：pending
    session-a → 入队时间 1000
    session-b → 入队时间 2000

String KV：dispatch:session-a
    → 调度状态、任务 ID、加密输入等 JSON
```

读取时先按 score 获取 member，再根据 ID 定位详情。也可以将 JSON 序列化为 member，但 JSON 内容变化会产生不同成员，不能自然地以其中某个业务 ID 保证唯一性。

## 5. 与底层编码的关系

数据模型描述对外语义；底层编码描述这些数据在内存中的具体组织方式。采用 listpack 或跳表与字典，都不改变 member 唯一、score 可重复及排序规则。

相关内容见 [ZSet](ZSet.md)。
