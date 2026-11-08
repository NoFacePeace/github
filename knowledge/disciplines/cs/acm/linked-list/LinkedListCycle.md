# 环形链表

相关知识：[链表](../../data-structures/linked-list/LinkedList.md)。

## 1. 题目描述

[LeetCode 141：环形链表](https://leetcode.cn/problems/linked-list-cycle/)。

给定单向链表的头节点 `head`，判断链表中是否存在环。如果从某个节点出发，不断沿 `Next` 指针能够再次到达该节点，则链表中存在环。

存在环时返回 `true`，否则返回 `false`。

题目使用 `pos` 表示尾节点连接到的节点下标（从 0 开始），`pos = -1` 表示尾节点不连接任何节点。`pos` 仅用于描述输入，不作为函数参数传入。

进阶要求：使用 O(1) 的额外空间解决问题。

## 2. 示例

```text
输入：head = [3, 2, 0, -4], pos = 1
输出：true
解释：尾节点的 Next 指向下标 1 的节点，形成环。
```

```text
输入：head = [1, 2], pos = 0
输出：true
解释：尾节点的 Next 指向头节点，形成环。
```

```text
输入：head = [1], pos = -1
输出：false
解释：尾节点的 Next 为空，不存在环。
```

## 3. Go 实现

使用[快慢指针](../../algorithms/two-pointers/FastSlowPointers.md)：慢指针每轮前进一步，快指针每轮前进两步；若存在环，两者最终会相遇，否则快指针会到达链表末尾。下面的实现让快指针从头节点的后继开始，并在每轮移动前比较两个指针。

也可以使用[哈希表](../../data-structures/hash-table/HashTable.md)记录访问过的节点引用，遇到重复节点时即可判定有环。应比较节点身份，而非节点值；该方法需要 O(n) 辅助空间，快慢指针则只需要 O(1) 辅助空间。

```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func hasCycle(head *ListNode) bool {
    if head == nil {
        return false
    }
    slow := head
    fast := head.Next
    for slow != nil && fast != nil {
        if slow == fast {
            return true
        }
        slow = slow.Next
        if fast.Next == nil {
            return false
        }
        fast = fast.Next.Next
    }
    return false
}
```
