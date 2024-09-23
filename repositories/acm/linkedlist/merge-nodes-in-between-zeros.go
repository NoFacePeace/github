package linkedlist

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func mergeNodes(head *ListNode) *ListNode {
	h := &ListNode{}
	h.Next = head
	cur := h
	for head != nil {
		if head.Val != 0 {
			cur.Val += head.Val
		} else {
			if head.Next != nil {
				cur.Next = head
				cur = head
			}
		}
		head = head.Next
	}

	cur.Next = nil
	return h.Next
}
