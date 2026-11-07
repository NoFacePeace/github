# LRU 缓存

相关知识：[LRU 淘汰算法](../../algorithms/eviction/lru/LRU.md)。

## 1. 题目描述

[LeetCode 146：LRU 缓存](https://leetcode.cn/problems/lru-cache/)。

设计并实现一个满足 LRU（最近最少使用）淘汰规则的缓存，支持以下操作：

- `LRUCache(int capacity)`：以正整数 `capacity` 初始化缓存容量。
- `int get(int key)`：如果关键字存在，返回对应的值，并将其标记为最近使用；否则返回 `-1`。
- `void put(int key, int value)`：如果关键字存在，更新其值；否则插入新的键值对。写入后将该关键字标记为最近使用，如果缓存超出容量，则淘汰最久未使用的关键字。

`get` 和 `put` 操作的平均时间复杂度均须为 O(1)。

## 2. 示例

```text
LRUCache cache = new LRUCache(2);
cache.put(1, 1);
cache.put(2, 2);
cache.get(1);    // 返回 1，关键字 1 成为最近使用。
cache.put(3, 3); // 淘汰关键字 2。
cache.get(2);    // 返回 -1。
cache.put(4, 4); // 淘汰关键字 1。
cache.get(1);    // 返回 -1。
cache.get(3);    // 返回 3。
cache.get(4);    // 返回 4。
```

## 3. Go 实现

使用[哈希表](../../data-structures/hash-table/HashTable.md)保存关键字到节点的映射，通过循环[双向链表](../../data-structures/linked-list/DoublyLinkedList.md)维护使用顺序：头节点为最近使用，尾节点为最久未使用。

```go
type LRUCache struct {
    head *node
    m map[int]*node
    cap int
}

type node struct {
    last *node
    next *node
    val int
    key int
}

func Constructor(capacity int) LRUCache {
    return LRUCache{
        cap: capacity,
        m: map[int]*node{},
    }
}

func (this *LRUCache) Get(key int) int {
    n, ok := this.m[key]
    if !ok {
        return -1
    }
    this.move(n)
    return n.val
}

func (this *LRUCache) Put(key int, value int) {
    if n, ok := this.m[key]; ok {
        n.val = value
        this.move(n)
        return
    }

    n := &node{key: key, val: value}
    n.last, n.next = n, n
    this.m[key] = n
    this.move(n)
    if len(this.m) > this.cap {
        this.remove()
    }
}

func (this *LRUCache) move(n *node) {
    if n == this.head {
        return
    }
    last := n.last
    next := n.next
    last.next = next
    next.last = last
    first := this.head
    if first == nil {
        this.head = n
        return
    }
    end := first.last
    n.next = first
    n.last = end
    end.next = n
    first.last = n
    this.head = n
}

func (this *LRUCache) remove() {
    end := this.head.last
    last := end.last
    this.head.last = last
    last.next = this.head
    delete(this.m, end.key)
}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
```
