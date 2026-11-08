# 删除链表的倒数第 N 个结点

相关知识：[链表](../../data-structures/linked-list/LinkedList.md)。

## 1. 题目描述

[LeetCode 19：删除链表的倒数第 N 个结点](https://leetcode.cn/problems/remove-nth-node-from-end-of-list/)。

给定单向链表的头节点 `head`，删除链表的倒数第 `n` 个节点，并返回删除后链表的头节点。

倒数第 1 个节点是尾节点。题目保证 `1 <= n <= 链表长度`；如果删除的是头节点，需要返回新的头节点。

进阶要求：使用一趟扫描实现。

## 2. 示例

```text
输入：head = [1, 2, 3, 4, 5], n = 2
输出：[1, 2, 3, 5]
解释：倒数第 2 个节点为值为 4 的节点，将其删除。
```

```text
输入：head = [1], n = 1
输出：[]
解释：删除唯一的节点后，链表为空。
```

```text
输入：head = [1, 2], n = 1
输出：[1]
```

## 3. Go 实现

使用[双指针](../../algorithms/two-pointers/TwoPointers.md)保持固定间隔：两个指针从哨兵节点出发，让 `next` 先前进 `n` 步，再让它与 `firt` 同速前进。当 `next` 到达尾节点时，`firt` 恰好位于待删除节点的前一个位置，修改链接即可完成删除。哨兵节点用于统一处理删除头节点的情况。

```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func removeNthFromEnd(head *ListNode, n int) *ListNode {
    tmp := &ListNode{}
    tmp.Next = head
    next := tmp
    for i := 0; i < n; i++ {
        next = next.Next
    }
    firt := tmp
    for next.Next != nil {
        firt = firt.Next
        next = next.Next
    }
    firt.Next = firt.Next.Next
    return tmp.Next
}
```
