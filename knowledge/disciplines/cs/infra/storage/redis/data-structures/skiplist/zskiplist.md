---
color: "#5B6CDA"
---

# zskiplist

`zskiplist` 是 Redis 跳表的整体管理结构，保存头尾节点、元素数量和当前有效层数。它通过 `header` 和 `tail` 指针连接 [zskiplistNode](zskiplistNode.md) 节点；节点之间通过各层的前向指针连接。

## 1. 结构定义

学习基线：Redis 7.2.16。

```c
typedef struct zskiplist {
    struct zskiplistNode *header, *tail;
    unsigned long length;
    int level;
} zskiplist;
```

| 字段 | 作用 |
| --- | --- |
| `header` | 头哨兵节点，不保存业务元素；它的各层前向指针提供查找入口。 |
| `tail` | 指向最后一个业务节点；空表时为 `NULL`，可作为倒序遍历入口。 |
| `length` | 业务节点数量，不包含头哨兵。 |
| `level` | 当前有效层数，至少为 1；不等于每个节点的层高。 |

## 2. 整体连接关系

```text
zskiplist
  length = 5
  level  = 3
  header ──→ 头哨兵
              第 2 层：head → 10 ─────────────────→ 50 → NULL
              第 1 层：head → 10 ───────→ 30 ─────→ 50 → NULL
              第 0 层：head → 10 → 20 → 30 → 40 → 50 → NULL
  tail ─────────────────────────────────────────→ 50
```

图中的数字代表分数，假设每个节点分数不同。同一个节点在多层出现，但内存中只保存一份节点；节点内部的 `level[]` 数组保存它参与各层的前向指针和跨度。

底层连接全部业务节点，业务节点还通过 `backward` 指针指向底层前驱；第一个业务节点的 `backward` 为 `NULL`。上层只连接部分节点，用于跨过一段元素、加速定位。

## 3. 初始化与维护

- `zslCreate` 创建头哨兵，为它分配最大层数的索引空间，并将各层前向指针置为 `NULL`。
- 空表的 `length` 为 0、`level` 为 1、`tail` 为 `NULL`。
- 插入时增加 `length`；新节点层高超过当前有效层数时，提高 `level`；插入尾部时更新 `tail`。
- 删除时减少 `length`；最高层变空时降低 `level`，但至少保留一层；删除尾节点时更新 `tail`。

## 4. 源码入口

- [server.h](https://github.com/redis/redis/blob/7.2.16/src/server.h)：`zskiplist` 与 `zskiplistNode` 的定义。
- [t_zset.c](https://github.com/redis/redis/blob/7.2.16/src/t_zset.c)：`zslCreate`、`zslInsert`、`zslDeleteNode` 和 `zslFree`。

[返回跳表目录](Skiplist.md)
