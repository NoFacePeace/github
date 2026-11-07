# 合并两个有序链表

相关知识：[有序链表](../../data-structures/linked-list/SortedLinkedList.md)。

## 1. 题目描述

[LeetCode 21：合并两个有序链表](https://leetcode.cn/problems/merge-two-sorted-lists/)。

给定两个按非递减顺序排列的单向链表的头节点 `list1` 和 `list2`，将两个链表合并为一个按非递减顺序排列的链表，并返回合并后链表的头节点。

新链表应通过拼接原有两个链表的节点组成。

## 2. 示例

```text
输入：list1 = [1, 2, 4], list2 = [1, 3, 4]
输出：[1, 1, 2, 3, 4, 4]
```

```text
输入：list1 = [], list2 = []
输出：[]
解释：两个链表均为空，合并后仍为空链表。
```

```text
输入：list1 = [], list2 = [0]
输出：[0]
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
func mergeTwoLists(list1 *ListNode, list2 *ListNode) *ListNode {
    if list1 == nil {
        return list2
    }
    if list2 == nil {
        return list1
    }
    var head *ListNode
    var next *ListNode
    for list1 != nil && list2 != nil {
        cur := list1
        if list1.Val > list2.Val {
            cur = list2
            list2 = list2.Next
        } else {
            list1 = list1.Next
        }
        if head == nil {
            head = cur
            next = head
            continue
        }
        next.Next = cur
        next = cur
    }
    if list1 != nil {
        next.Next = list1
    }
    if list2 != nil {
        next.Next = list2
    }
    return head
}
```
