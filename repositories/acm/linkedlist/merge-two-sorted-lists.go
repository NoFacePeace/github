package linkedlist

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func mergeTwoLists(list1 *ListNode, list2 *ListNode) *ListNode {
	head := &ListNode{}
	last := head
	for list1 != nil && list2 != nil {
		if list1.Val < list2.Val {
			last.Next = list1
			last = last.Next
			list1 = list1.Next
		} else {
			last.Next = list2
			last = last.Next
			list2 = list2.Next
		}
	}
	if list1 != nil {
		last.Next = list1
	}
	if list2 != nil {
		last.Next = list2
	}
	return head.Next
}
