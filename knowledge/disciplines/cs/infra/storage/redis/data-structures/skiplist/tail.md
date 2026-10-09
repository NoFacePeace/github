---
color: "#5B6CDA"
---

# tail

`tail` 是 `zskiplist` 中指向最后一个业务节点的指针，用于直接访问跳表末尾并开始倒序遍历。学习基线为 Redis 7.2.16。

## 1. 字段与指向

```c
struct zskiplistNode *tail;
```

`tail` 指向已有的 `zskiplistNode` 业务节点，不额外创建尾哨兵。空表时为 `NULL`；非空时指向底层有序链表的最后一个节点。

```text
底层：head → 10 → 20 → 30 → NULL
                           ↑
tail ──────────────────────┘
```

图中数字代表分数，假设各节点分数不同。最后一个节点具有最高分数；最高分相同时，位于末尾的是成员字节序最大的节点。尾节点层高不一定最高。

## 2. 倒序遍历

从 `tail` 出发，沿业务节点的 `backward` 指针可以依次访问前驱：

```text
tail → 30 → 20 → 10 → NULL
           沿 backward
```

第一个业务节点的 `backward` 为 `NULL`，不会指向头哨兵。从末尾读取 k 个元素的遍历成本为 O(k)，无须先从头遍历到尾。

## 3. 初始化与维护

- `zslCreate` 将 `tail` 初始化为 `NULL`。
- `zslInsert` 插入新节点后，若该节点没有底层后继，就将 `tail` 更新为新节点；向空表插入第一个节点也属于这种情况。
- `zslDeleteNode` 删除尾节点时，将 `tail` 更新为被删除节点的 `backward`。
- 删除唯一的业务节点后，`tail` 恢复为 `NULL`，不指向头哨兵。
- 分数变化需要移除并重新插入节点时，通过删除与插入过程维护 `tail`。

## 4. 源码入口

- [server.h](https://github.com/redis/redis/blob/7.2.16/src/server.h)：`zskiplist.tail` 和节点的 `backward` 字段。
- [t_zset.c](https://github.com/redis/redis/blob/7.2.16/src/t_zset.c)：`zslCreate`、`zslInsert`、`zslDeleteNode` 和 `zslUpdateScore`。

[返回整体结构](zskiplist.md)
