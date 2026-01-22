package search

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func maxLevelSum(root *TreeNode) int {
	m := map[int]int{}
	mx := 1
	var f func(root *TreeNode, level int)
	f = func(root *TreeNode, level int) {
		if root == nil {
			return
		}
		mx = max(mx, level)
		m[level] += root.Val
		f(root.Left, level+1)
		f(root.Right, level+1)
	}
	f(root, 1)
	ans := m[1]
	level := 1
	for i := 1; i <= mx; i++ {
		if m[i] > ans {
			ans = m[i]
			level = i
		}
	}
	return level
}
