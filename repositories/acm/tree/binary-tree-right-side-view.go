package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func rightSideView(root *TreeNode) []int {
	inc := -1
	arr := []int{}
	var f func(root *TreeNode, h int)
	f = func(root *TreeNode, h int) {
		if root == nil {
			return
		}
		if h > inc {
			inc = h
			arr = append(arr, root.Val)
		}
		f(root.Right, h+1)
		f(root.Left, h+1)
	}
	f(root, 0)
	return arr
}
