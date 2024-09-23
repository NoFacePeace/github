package tree

import "math"

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func maxPathSum(root *TreeNode) int {
	mx := math.MinInt
	var f func(root *TreeNode) int
	f = func(root *TreeNode) int {
		if root == nil {
			return 0
		}
		left := f(root.Left)
		right := f(root.Right)
		cur := root.Val
		cur = max(cur, root.Val+left)
		cur = max(cur, root.Val+right)
		cur = max(cur, root.Val+left+right)
		if cur > mx {
			mx = cur
		}
		return max(max(root.Val+left, root.Val+right), root.Val)
	}
	ret := f(root)
	return max(ret, mx)
}
