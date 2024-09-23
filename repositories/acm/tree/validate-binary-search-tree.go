package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func isValidBST(root *TreeNode) bool {
	ans := true
	var pre *TreeNode
	var f func(*TreeNode)
	f = func(root *TreeNode) {
		if root == nil {
			return
		}
		f(root.Left)
		if pre == nil {
			pre = root
		} else {
			if pre.Val >= root.Val {
				ans = false
				return
			}
			pre = root
		}
		f(root.Right)
	}
	f(root)
	return ans
}
