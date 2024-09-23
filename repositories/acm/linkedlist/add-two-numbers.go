package linkedlist

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	bit := 0
	head := l1
	last := l1
	for l1 != nil && l2 != nil {
		l1.Val += l2.Val + bit
		if l1.Val >= 10 {
			bit = 1
			l1.Val %= 10
		} else {
			bit = 0
		}
		last = l1
		l1 = l1.Next
		l2 = l2.Next
	}
	if l1 == nil {
		last.Next = l2
	}
	for l2 != nil {
		l2.Val += bit
		if l2.Val < 10 {
			bit = 0
			break
		}
		bit = 1
		l2.Val %= 10
		last = l2
		l2 = l2.Next
	}
	for l1 != nil {
		l1.Val += bit
		if l1.Val < 10 {
			bit = 0
			break
		}
		bit = 1
		l1.Val %= 10
		last = l1
		l1 = l1.Next
	}
	if bit == 1 {
		last.Next = &ListNode{Val: 1}
	}
	return head
}
