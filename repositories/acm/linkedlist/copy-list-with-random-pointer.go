package linkedlist

type Node struct {
	Val    int
	Next   *Node
	Random *Node
}

func copyRandomList(head *Node) *Node {
	if head == nil {
		return nil
	}
	m := map[*Node]*Node{}
	f := func(node *Node) *Node {
		if node == nil {
			return nil
		}
		if n, ok := m[node]; ok {
			return n
		}
		n := &Node{}
		n.Val = node.Val
		m[node] = n
		return n
	}
	copy := &Node{}
	last := copy
	for head != nil {
		node := f(head)
		node.Random = f(head.Random)
		last.Next = node
		last = last.Next
		head = head.Next
	}
	return copy.Next
}
