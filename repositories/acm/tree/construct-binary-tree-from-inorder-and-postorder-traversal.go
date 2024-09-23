package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func buildTreeI(inorder []int, postorder []int) *TreeNode {
	n := len(postorder)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return &TreeNode{
			Val: postorder[0],
		}
	}
	num := postorder[n-1]
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
	node.Left = buildTree(inorder[:inc], postorder[:inc])
	node.Right = buildTree(inorder[inc+1:], postorder[inc:n-1])
	return node
}
