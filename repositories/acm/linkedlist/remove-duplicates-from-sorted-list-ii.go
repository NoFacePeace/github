package linkedlist

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func deleteDuplicates(head *ListNode) *ListNode {
	if head == nil {
		return nil
	}
	h := &ListNode{}
	next := h
	last := head
	head = head.Next
	inc := 1
	for head != nil {
		if head.Val == last.Val {
			inc++
		} else if inc == 1 {
			next.Next = last
			last.Next = nil
			next = next.Next
		} else {
			inc = 1
		}
		last = head
		head = head.Next
	}
	if inc == 1 {
		next.Next = last
	}
	return h.Next
}
