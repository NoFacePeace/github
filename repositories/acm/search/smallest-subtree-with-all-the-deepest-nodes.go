package search

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func subtreeWithAllDeepest(root *TreeNode) *TreeNode {
	var dfs func(root *TreeNode, d int) int
	ans := root
	depth := 0
	dfs = func(root *TreeNode, d int) int {
		if root == nil {
			return 0
		}
		l := dfs(root.Left, d+1)
		r := dfs(root.Right, d+1)
		if l != r {
			return max(l, r) + 1
		}
		if d+l >= depth {
			ans = root
			depth = d + l
		}
		return l + 1
	}
	dfs(root, 0)
	return ans
}
