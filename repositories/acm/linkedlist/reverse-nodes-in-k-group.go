package linkedlist

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func reverseKGroup(head *ListNode, k int) *ListNode {
	if head == nil {
		return nil
	}
	h := &ListNode{}
	h.Next = head
	pre := h
	inc := 1
	for head != nil && head.Next != nil {
		next := head.Next
		head.Next = next.Next
		next.Next = pre.Next
		pre.Next = next
		inc++
		if inc == k {
			inc = 1
			pre = head
			head = head.Next
		}
	}
	if inc == 1 {
		return h.Next
	}
	head = pre.Next
	for head.Next != nil {
		next := head.Next
		head.Next = next.Next
		next.Next = pre.Next
		pre.Next = next
	}
	return h.Next
}
