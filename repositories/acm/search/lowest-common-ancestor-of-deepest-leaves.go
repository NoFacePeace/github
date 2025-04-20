package search

// https://leetcode.cn/problems/lowest-common-ancestor-of-deepest-leaves/?envType=daily-question&envId=2025-04-04

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func lcaDeepestLeaves(root *TreeNode) *TreeNode {
	var dfs func(root *TreeNode, d int) int
	var ans *TreeNode
	mx := 0
	dfs = func(root *TreeNode, d int) int {
		if root == nil {
			return d
		}
		l := dfs(root.Left, d+1)
		r := dfs(root.Right, d+1)
		if l != r {
			return max(l, r)
		}
		if l >= mx {
			ans = root
			mx = l
		}
		return l
	}
	dfs(root, 0)
	return ans
}
