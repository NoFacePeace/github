package bitwise

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func sumRootToLeaf(root *TreeNode) int {
	ans := 0
	var dfs func(root *TreeNode, num int)
	dfs = func(root *TreeNode, num int) {
		if root == nil {
			return
		}
		num = num << 1
		num += root.Val
		dfs(root.Left, num)
		dfs(root.Right, num)
		if root.Left == nil && root.Right == nil {
			ans += num
		}
	}
	dfs(root, 0)
	return ans
}
