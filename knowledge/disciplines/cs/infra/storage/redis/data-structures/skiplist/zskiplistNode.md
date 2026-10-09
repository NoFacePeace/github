---
color: "#5B6CDA"
---

# zskiplistNode

`zskiplistNode` 是 Redis 跳表的节点结构，保存成员字符串、分数、底层后退指针，以及各层的前向指针和跨度。

## 1. 结构定义

学习基线：Redis 7.2.16。

```c
typedef struct zskiplistNode {
    sds ele;
    double score;
    struct zskiplistNode *backward;
    struct zskiplistLevel {
        struct zskiplistNode *forward;
        unsigned long span;
    } level[];
} zskiplistNode;
```

| 字段 | 作用 |
| --- | --- |
| `ele` | 保存成员字符串，使用 SDS。 |
| `score` | 保存分数，节点先按分数排序，分数相同时按成员字节序排序。 |
| `backward` | 指向底层前驱；第一个业务节点为 `NULL`，用于倒序遍历。 |
| `level[i].forward` | 指向第 i 层的后继节点；没有后继时为 `NULL`。 |
| `level[i].span` | 记录该层连接跨越的底层元素数量，用于排名计算和按排名定位。 |

## 2. 层高与内存布局

`level[]` 是柔性数组成员。创建节点时，Redis 按层高分配相应数量的索引项：

```c
sizeof(zskiplistNode) + level * sizeof(struct zskiplistLevel)
```

一个层高为 3 的节点，包含 `level[0]`、`level[1]`、`level[2]` 三项。节点只保存一份成员和分数，各层通过不同的前向指针连接同一批节点。

```text
节点
├── ele
├── score
├── backward ──→ 底层前驱
├── level[0]：forward、span
├── level[1]：forward、span
└── level[2]：forward、span
```

业务节点的层高由 `zslRandomLevel` 随机生成，至少为 1，最大为 `ZSKIPLIST_MAXLEVEL`。Redis 7.2.16 中最大层数为 32，每次继续增加一层的概率约为 1/4。头哨兵直接分配最大层数，不保存业务成员。

## 3. span 的含义

当 `forward` 指向业务节点时，`span` 表示起点到终点的底层排名差：不包含起点，包含终点。

```text
底层：10 → 20 → 30 → 40
高层：10 ─────────→ 40
          span = 3
```

从 10 到 40 跨过的底层元素为 20、30、40，因此跨度为 3。查找时累加经过连接的跨度，可以计算排名。

当 `forward` 为 `NULL` 时，跨度记录该节点之后剩余的底层元素数量，因此高层的跨度不一定为 0。尾节点之后没有元素，跨度为 0。

## 4. 指针与跨度维护

- 插入节点时，调整各层前向连接并拆分跨度；新节点未参与的高层，也需要增加跨过插入位置的连接跨度。
- 删除节点时，合并直接连接目标节点的跨度，并缩短跨过目标位置的高层跨度。
- 底层前后关系变化时，更新后继节点的 `backward`；尾部变化还需要更新整体结构的 `tail`。
- 分数更新导致排序位置变化时，节点需要移除后重新插入。

## 5. 源码入口

- [server.h](https://github.com/redis/redis/blob/7.2.16/src/server.h)：`zskiplistNode`、`ZSKIPLIST_MAXLEVEL` 和 `ZSKIPLIST_P`。
- [t_zset.c](https://github.com/redis/redis/blob/7.2.16/src/t_zset.c)：`zslCreateNode`、`zslRandomLevel`、`zslInsert`、`zslDeleteNode` 和 `zslUpdateScore`。

[返回跳表目录](Skiplist.md)
