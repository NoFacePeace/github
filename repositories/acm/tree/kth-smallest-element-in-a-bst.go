package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func kthSmallest(root *TreeNode, k int) int {
	ans := 0
	inc := 0
	var f func(*TreeNode)
	f = func(root *TreeNode) {
		if root == nil {
			return
		}
		f(root.Left)
		inc++
		if inc == k {
			ans = root.Val
			return
		}
		f(root.Right)
	}
	f(root)
	return ans
}
