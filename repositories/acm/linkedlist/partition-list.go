package linkedlist

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func partition(head *ListNode, x int) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}
	h := &ListNode{}
	h.Next = head
	pre := h
	for pre.Next != nil && pre.Next.Val < x {
		pre = pre.Next
	}
	last := pre
	head = last.Next
	for head != nil {
		if head.Val < x {
			last.Next = head.Next
			head.Next = pre.Next
			pre.Next = head
			pre = pre.Next
			head = last.Next
			continue
		}
		last = head
		head = head.Next
	}
	return h.Next
}
