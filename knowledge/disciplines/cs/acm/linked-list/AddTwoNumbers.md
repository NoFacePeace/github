# 两数相加

相关知识：[链表](../../data-structures/linked-list/LinkedList.md)。

## 1. 题目描述

[LeetCode 2：两数相加](https://leetcode.cn/problems/add-two-numbers/)。

给定两个非空单向链表 `l1` 和 `l2`，分别表示两个非负整数。每个节点保存一位十进制数字，数字按逆序存储，即头节点表示个位。

将这两个整数相加，并以相同的逆序链表形式返回结果。

除数字 0 本身外，输入整数不包含前导零。两个链表的长度可以不同，每个节点的值都在 0 到 9 之间。

## 2. 示例

```text
输入：l1 = [2, 4, 3], l2 = [5, 6, 4]
输出：[7, 0, 8]
解释：两个链表分别表示 342 和 465，342 + 465 = 807，结果逆序存储为 [7, 0, 8]。
```

```text
输入：l1 = [0], l2 = [0]
输出：[0]
```

```text
输入：l1 = [9, 9, 9, 9, 9, 9, 9], l2 = [9, 9, 9, 9]
输出：[8, 9, 9, 9, 0, 0, 0, 1]
解释：9999999 + 9999 = 10009998。
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
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
    head := l1
    last := l1
    if l1 == nil {
        head = l2
        last = l2
    }
    carry := 0
    for l1 != nil && l2 != nil {
        val := carry + l1.Val + l2.Val
        l1.Val = val % 10
        carry = val / 10
        last = l1
        l1 = l1.Next
        l2 = l2.Next
    }
    for l1 != nil {
        val := l1.Val + carry
        l1.Val = val % 10
        carry = val / 10
        last = l1
        l1 = l1.Next
    }
    for l2 != nil {
        if last != l2 {
            last.Next = l2
        }
        val := l2.Val + carry
        l2.Val = val % 10
        carry = val / 10
        last = l2
        l2 = l2.Next
    }
    if carry != 0 {
        last.Next = &ListNode{
            Val: 1,
        }
    }
    return head
}
```
