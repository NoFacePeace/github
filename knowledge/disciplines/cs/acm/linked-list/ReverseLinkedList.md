# 反转链表

相关知识：[链表](../../data-structures/linked-list/LinkedList.md)。

## 1. 题目描述

[LeetCode 206：反转链表](https://leetcode.cn/problems/reverse-linked-list/)。

给定单向链表的头节点 `head`，反转该链表，并返回反转后链表的头节点。

## 2. 示例

```text
输入：head = [1, 2, 3, 4, 5]
输出：[5, 4, 3, 2, 1]
```

```text
输入：head = [1, 2]
输出：[2, 1]
```

```text
输入：head = []
输出：[]
解释：空链表反转后仍为空链表。
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
func reverseList(head *ListNode) *ListNode {
    var last *ListNode
    for head != nil {
        tmp := head.Next
        head.Next = last
        last = head
        head = tmp
    }
    return last
}
```
