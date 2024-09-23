package linkedlist

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func removeNthFromEnd(head *ListNode, n int) *ListNode {
	inc := 0
	var f func(head *ListNode) *ListNode
	f = func(head *ListNode) *ListNode {
		if head == nil {
			return nil
		}
		next := f(head.Next)
		inc++
		if inc == n {
			return next
		}
		head.Next = next
		return head
	}
	head = f(head)
	return head
}
