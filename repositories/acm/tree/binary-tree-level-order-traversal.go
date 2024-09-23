package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func levelOrder(root *TreeNode) [][]int {
	q := []*TreeNode{root}
	arr := [][]int{}
	for len(q) > 0 {
		tmp := q
		q = nil
		sub := []int{}
		for _, v := range tmp {
			if v == nil {
				continue
			}
			sub = append(sub, v.Val)
			q = append(q, v.Left, v.Right)
		}
		if len(sub) == 0 {
			continue
		}
		arr = append(arr, sub)
	}
	return arr
}
