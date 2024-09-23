package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func sumNumbers(root *TreeNode) int {
	ans := 0
	var f func(root *TreeNode, sum int)
	f = func(root *TreeNode, sum int) {
		if root == nil {
			return
		}
		if root.Left == nil && root.Right == nil {
			ans += sum*10 + root.Val
			return
		}
		f(root.Right, sum*10+root.Val)

		f(root.Left, sum*10+root.Val)
	}
	f(root, 0)
	return ans
}
