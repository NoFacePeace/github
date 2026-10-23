---
color: "#5B6CDA"
---

# NX

`NX` 表示仅在目标不存在时执行相应操作。目标是什么由命令决定：`SET` 检查 key，`ZADD` 检查集合中的 member。

## 1. SET 中的 NX

```redis
SET task-lock token-a NX PX 30000
SET task-lock token-b NX PX 30000
```

假设 key 原先不存在，且两次调用之间没有过期或删除：

- 第一条返回 `OK`，写入 `token-a` 并设置 30 秒有效期。
- 第二条返回空响应，表示没有写入；原值和过期时间不变。

这是未使用 `GET` 选项时的常见返回规则。`SET` 组合 `GET` 时返回语义不同，不能照搬“返回 `OK` 才成功”的判断方式。

`NX` 本身不设置有效期；不指定过期策略时，新建 key 默认不会自动过期。

## 2. 原子条件写入

```text
客户端 A：SET task-lock token-a NX → 成功
客户端 B：SET task-lock token-b NX → 条件不满足
```

对同一 Redis 节点上的同一个 key，存在性检查和写入在一条命令内完成，其他客户端无法在两者之间插入操作。

不要在客户端用 `GET` 检查后再无条件 `SET`，否则两个客户端可能都读到 key 不存在，然后分别写入并认为自己成功。

## 3. ZADD 中的 NX

```redis
ZADD ranking 10 alice
ZADD ranking NX 20 alice 30 bob
```

假设 `ranking` 初始不存在，第二条只新增 `bob`，不会修改已有 `alice` 的分数。因此最终为 `alice → 10`、`bob → 30`，第二条返回新增数量 `1`。

这里即使 ZSet 的 key 已存在，也可以添加新的 member。`NX` 检查的是 member 是否存在，而不是整个集合是否存在。

## 4. 命令差异与限制

| 命令 | NX 的判断目标 | 条件满足时的动作 |
| --- | --- | --- |
| `SET key value NX` | key 不存在。 | 创建 key 并写值。 |
| `ZADD key NX score member` | member 不存在。 | 添加该成员与分数。 |
| `EXPIRE key seconds NX` | key 没有过期时间。 | 为已有 key 设置过期时间；Redis 7.0 起支持。 |

所以不能把 `NX` 理解为任何命令下都表示“key 不存在”。使用时必须查看对应命令定义。

`SET` 中 `NX` 与 `XX` 互斥；`ZADD` 中 `NX` 不能与 `XX`、`GT`、`LT` 组合。其他命令的选项约束由各自语法决定。

## 5. 临时占用与锁

`SET ... NX PX ...` 可以用作带期限的资源占用：NX 决定是否获得占用，PX 限制占用时间。它不自动提供公平排队、续约、身份校验释放或严格的跨故障互斥。

锁过期后，其他实例可能获取同一个 key；旧实例释放时必须比较自己的 token，再原子删除，不能无条件 `DEL`。资源端若需拒绝旧实例继续写入，还需要独立的有效性保护。
