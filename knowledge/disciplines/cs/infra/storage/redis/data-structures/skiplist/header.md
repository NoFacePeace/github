---
color: "#5B6CDA"
---

# header

`header` 是 [zskiplist](zskiplist.md) 中指向头哨兵节点的指针。头哨兵使用 [zskiplistNode](zskiplistNode.md) 结构，提供各层的起始连接，不保存业务元素，也不计入 `length`。

## 1. 字段与节点

```c
struct zskiplistNode *header;
```

`header` 字段本身是一个指针，它指向单独分配的节点。该节点的 `level[]` 数组保存各层的前向指针与跨度，并不是每层各有一个头节点。

```text
header → 头哨兵
          level[2].forward → 第 2 层首节点
          level[1].forward → 第 1 层首节点
          level[0].forward → 第一个业务节点
```

某层没有业务节点时，该层的 `forward` 为 `NULL`。不同层的首节点可能是同一个业务节点，也可能不同。

## 2. 初始化

学习基线：Redis 7.2.16。`zslCreate` 使用以下调用创建头哨兵：

```c
zsl->header = zslCreateNode(ZSKIPLIST_MAXLEVEL, 0, NULL);
```

- 分配最大层数的索引空间，此版本的 `ZSKIPLIST_MAXLEVEL` 为 32。
- `ele` 为 `NULL`，`score` 初始化为 0；该分数不参与业务节点排序。
- 各层的 `forward` 初始化为 `NULL`，`span` 初始化为 0。
- `backward` 为 `NULL`。

头哨兵分配 32 层，不代表跳表当前使用 32 层。空表的有效层数 `zskiplist.level` 为 1，后续查找从当前有效最高层开始。

## 3. 查找与修改的入口

查找从头哨兵开始，在当前层沿前向指针前进；继续前进会超过目标时，下降一层。头哨兵为所有有效层提供统一入口。

插入到某层最前面时，该层的前驱就是头哨兵，只需修改它的前向指针与跨度。删除首节点时也通过相同方式调整连接。

头哨兵不是底层后退遍历的终点节点：第一个业务节点的 `backward` 为 `NULL`，而不是指向头哨兵。

## 4. 跨度与生命周期

- 当头哨兵的某层前向指针指向业务节点时，`span` 等于该节点从 1 开始的底层排名。
- 插入和删除会更新相关层的前向连接及跨度。
- 删除全部业务节点后，头哨兵仍然保留，跳表可以继续插入元素。
- 释放整个跳表时，`zslFree` 会释放头哨兵及业务节点。

## 5. 源码入口

- [server.h](https://github.com/redis/redis/blob/7.2.16/src/server.h)：`zskiplist.header`、`zskiplistNode` 和最大层数定义。
- [t_zset.c](https://github.com/redis/redis/blob/7.2.16/src/t_zset.c)：`zslCreate`、`zslInsert`、`zslDeleteNode` 和 `zslFree`。

[返回整体结构](zskiplist.md)
