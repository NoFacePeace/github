package tree

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func lowestCommonAncestor(root, p, q *TreeNode) *TreeNode {
	var dfs func(node *TreeNode) bool
	dfs = func(node *TreeNode) bool {
		if node == nil {
			return false
		}
		left := dfs(node.Left)
		right := dfs(node.Right)
		if left && right {
			root = node
			return true
		}
		if left && (node == p || node == q) {
			root = node
			return true
		}
		if right && (node == p || node == q) {
			root = node
			return true
		}
		return node == p || node == q || left || right
	}
	dfs(root)
	return root
}
