# 移到头部

`moveToHead(node)` 将 LRU 缓存中已有的数据节点移到头哨兵之后，表示该节点最近被使用。查询命中和更新已有键时，都会用到这个操作。

## 1. 前置条件

- 链表使用头尾哨兵 `head`、`tail`，从头到尾按“最近使用 → 最久未使用”排列。
- `node` 是当前链表中的数据节点，不能是空指针、哨兵或其他链表的节点。
- 节点在哈希表中的映射已经存在。

## 2. 操作流程

1. 如果 `node` 已经是 `head.Next`，直接结束。
2. 保存节点的前驱和后继，将前后两个节点重新连接，摘除 `node`。
3. 保存摘除后头哨兵的后继节点 `first`。
4. 将 `node` 插入 `head` 与 `first` 之间，同时维护前驱和后继链接。

伪代码如下：

```text
moveToHead(node):
    if head.Next == node:
        return

    prev = node.Prev
    next = node.Next
    prev.Next = next
    next.Prev = prev

    first = head.Next
    node.Prev = head
    node.Next = first
    head.Next = node
    first.Prev = node
```

## 3. 示例

```text
移动前：head ⇄ C ⇄ B ⇄ A ⇄ tail
操作：moveToHead(B)
移动后：head ⇄ B ⇄ C ⇄ A ⇄ tail
```

B 成为最近使用的节点，其余节点的相对顺序不变。

## 4. 注意事项与复杂度

- 必须先摘除再插入，直接插入可能破坏原来的链接或形成环。
- 节点的键、值和身份均不改变，哈希表映射与节点数量保持不变。
- 移动尾部的数据节点时，摘除步骤会同步更新 `tail.Prev`。
- 只有一个数据节点时，该节点已经在头部，直接结束即可。
- 并发实现中，摘除与重新插入应处于同一次同步保护之内。

操作只调整固定数量的指针，时间复杂度和额外空间复杂度均为 O(1)。
