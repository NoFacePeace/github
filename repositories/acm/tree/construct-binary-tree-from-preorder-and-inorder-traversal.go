package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func buildTree(preorder []int, inorder []int) *TreeNode {
	n := len(preorder)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return &TreeNode{
			Val: preorder[0],
		}
	}
	num := preorder[0]
	inc := 0
	for _, v := range inorder {
		if v == num {
			break
		}
		inc++
	}
	node := &TreeNode{
		Val: num,
	}
	node.Left = buildTree(preorder[1:inc+1], inorder[:inc])
	node.Right = buildTree(preorder[inc+1:], inorder[inc+1:])
	return node
}
