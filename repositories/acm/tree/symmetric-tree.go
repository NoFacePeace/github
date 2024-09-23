package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func isSymmetric(root *TreeNode) bool {
	if root == nil {
		return true
	}
	if root.Left == nil && root.Right == nil {
		return true
	}
	if root.Left == nil {
		return false
	}
	if root.Right == nil {
		return false
	}
	if root.Left.Val != root.Right.Val {
		return false
	}
	root.Left.Right, root.Right.Right = root.Right.Right, root.Left.Right
	return isSymmetric(root.Left) && isSymmetric(root.Right)
}
