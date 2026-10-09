---
color: "#5B6CDA"
---

# level[]

`level[]` 是跳表节点中的柔性数组，每个元素都是一个 [zskiplistLevel](zskiplistLevel.md)，保存节点在对应层的前向指针和跨度。

## 1. 数组与元素

```c
struct zskiplistLevel {
    struct zskiplistNode *forward;
    unsigned long span;
} level[];
```

`level[]` 组织节点参与的所有层，`level[i]` 则是第 i 层的索引项：

```text
level[]
├── [0]：zskiplistLevel → forward、span
├── [1]：zskiplistLevel → forward、span
└── [2]：zskiplistLevel → forward、span
```

`level[0]` 对应底层。层高为 h 的节点拥有 h 个索引项，有效下标为 0 到 h−1；不同节点的数组长度可以不同。

## 2. 柔性数组与内存分配

`level[]` 位于节点结构的末尾，声明时不指定长度。创建节点时，按层高为它分配额外空间：

```c
sizeof(zskiplistNode) + h * sizeof(struct zskiplistLevel)
```

数组元素与节点固定字段位于同一次分配的内存中。`level[]` 本身不是指向独立数组的指针，`sizeof(zskiplistNode)` 也不包含这些额外索引项的存储空间。

## 3. 与有效层数的区别

- 节点的 `level[]` 长度表示该节点的层高，决定它能参与哪些层。
- 整体结构的 `zskiplist.level` 是一个整数，表示整个跳表当前的有效层数。
- 头哨兵的 `level[]` 分配最大层数的空间，但查找只使用当前有效层。

## 4. 源码入口

- `server.h`：节点中的 `level[]` 与 `zskiplistLevel` 定义。
- `t_zset.c`：`zslCreateNode` 根据层高分配节点及数组空间。

[返回节点结构](zskiplistNode.md)
