package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func flatten(root *TreeNode) {
	if root == nil {
		return
	}
	flatten(root.Left)
	flatten(root.Right)
	if root.Left == nil {
		return
	}
	left := root.Left
	right := root.Right
	root.Right = root.Left
	root.Left = nil
	for left.Right != nil {
		left = left.Right
	}
	left.Right = right
}
