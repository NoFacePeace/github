package tree

type Node struct {
	Val   int
	Left  *Node
	Right *Node
	Next  *Node
}

func connect(root *Node) *Node {
	if root == nil {
		return nil
	}
	q := []*Node{root}
	for len(q) > 0 {
		tmp := q
		q = nil
		n := len(tmp)
		for i := 0; i < n; i++ {
			if i < n-1 {
				tmp[i].Next = tmp[i+1]
			}
			if tmp[i].Left != nil {
				q = append(q, tmp[i].Left)
			}
			if tmp[i].Right != nil {
				q = append(q, tmp[i].Right)
			}
		}
	}
	return root
}
