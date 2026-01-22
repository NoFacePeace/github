package search

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func maxProduct(root *TreeNode) int {
	sum := 0
	var f func(root *TreeNode)
	f = func(root *TreeNode) {
		if root == nil {
			return
		}
		sum += root.Val
		f(root.Left)
		f(root.Right)
	}
	f(root)
	ans := 0
	mod := int(1e9) + 7
	var compute func(root *TreeNode) int
	compute = func(root *TreeNode) int {
		if root == nil {
			return 0
		}
		l := compute(root.Left)
		ans = max(ans, (sum-l)*l)
		r := compute(root.Right)
		ans = max(ans, (sum-r)*r)
		return root.Val + l + r
	}

	compute(root)
	return ans % mod
}
