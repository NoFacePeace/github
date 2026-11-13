# K 个一组翻转链表

相关知识：[链表](../../data-structures/linked-list/LinkedList.md)。

## 1. 题目描述

[LeetCode 25：K 个一组翻转链表](https://leetcode.cn/problems/reverse-nodes-in-k-group/)。

给定单向链表的头节点 `head`，每 `k` 个节点为一组进行翻转，并返回修改后链表的头节点。

如果最后剩余的节点数量不足 `k` 个，则保持它们的原有顺序。不能只修改节点内部的值，必须实际调整节点之间的链接。

题目保证 `1 <= k <= 链表长度`。

进阶要求：使用 O(1) 的额外空间解决问题。

## 2. 示例

```text
输入：head = [1, 2, 3, 4, 5], k = 2
输出：[2, 1, 4, 3, 5]
解释：每两个节点翻转一次，最后一个节点保持不变。
```

```text
输入：head = [1, 2, 3, 4, 5], k = 3
输出：[3, 2, 1, 4, 5]
解释：前三个节点翻转，剩余两个节点不足一组，保持原有顺序。
```

## 3. Go 实现

```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func reverseKGroup(head *ListNode, k int) *ListNode {
    dummy := &ListNode{}
    dummy.Next = head
    var f func(cur *ListNode) *ListNode
    f = func(cur *ListNode) *ListNode {
        if cur == nil {
            return nil
        }
        next := cur
        for i := 1; i < k; i++ {
            next = next.Next
            if next == nil {
                return cur
            }
        }
        last := cur
        next = cur.Next
        for i := 1; i < k; i++ {
            tmp := next.Next
            next.Next = last
            last = next
            next = tmp
        }
        cur.Next = f(next)
        return last
    }
    dummy.Next = f(head)
    return dummy.Next
}
```
