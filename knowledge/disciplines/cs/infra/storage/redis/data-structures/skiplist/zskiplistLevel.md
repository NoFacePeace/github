---
color: "#5B6CDA"
---

# zskiplistLevel

`zskiplistLevel` 描述一个跳表节点在某一层的连接，包含前向指针 `forward` 和跨度 `span`。节点的 `level[]` 数组由这些结构组成，每个数组元素对应一层。

## 1. 结构定义

以下定义位于 `zskiplistNode` 内部：

```c
struct zskiplistLevel {
    struct zskiplistNode *forward;
    unsigned long span;
} level[];
```

| 字段 | 作用 |
| --- | --- |
| `forward` | 指向同一层的下一个节点；没有后继时为 `NULL`。 |
| `span` | 记录连接跨越的底层元素数量，用于计算排名和按排名定位。 |

它表示单个节点在一层上的索引项。整层的连接由多个节点各自的 `level[i].forward` 串接而成。

## 2. forward 与 span

```text
底层：10 → 20 → 30 → 40
高层：10 ─────────→ 40
          span = 3
```

假设节点 10 和 40 都参与第 1 层，则节点 10 的 `level[1].forward` 指向节点 40，`level[1].span` 为 3。

当 `forward` 指向业务节点时，跨度等于两节点的底层排名差，不包含起点、包含终点。从 10 到 40 计入 20、30、40 三个元素。

当 `forward` 为 `NULL` 时，跨度记录当前节点之后剩余的底层元素数量，因此高层的空指针不一定对应零跨度。尾节点之后没有元素，其跨度为 0。

## 3. 与节点层高的关系

- 层高为 h 的业务节点分配 h 个 `zskiplistLevel`，有效下标为 0 到 h−1。
- `level[0]` 对应底层，每个业务节点都参与；更高层只连接具有相应层高的节点。
- 同一节点各层的 `forward` 可以指向不同后继，各层的 `span` 也可以不同。
- 每层只有前向指针，底层后退指针 `backward` 保存在节点本身，不属于 `zskiplistLevel`。

## 4. 插入与删除时的维护

- 插入参与某层的新节点时，将原前向连接拆成两段，并分别计算跨度。
- 新节点未参与的高层，跨过插入位置的连接保持不变，跨度增加 1。
- 删除直接连接的节点时，合并前后两段连接，跨度为两段跨度之和减 1。
- 删除未参与某层的节点时，该层跨过删除位置的连接保持不变，跨度减少 1。

查找排名时，沿经过的前向连接累加跨度；按排名定位时，根据累积跨度判断是否继续前进。

## 5. 源码入口

- `server.h`：`zskiplistLevel` 和节点的 `level[]` 定义。
- `t_zset.c`：`zslInsert`、`zslDeleteNode`、`zslGetRank` 和 `zslGetElementByRank`。

[返回跳表目录](Skiplist.md)
