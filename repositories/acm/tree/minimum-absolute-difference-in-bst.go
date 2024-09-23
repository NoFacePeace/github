package tree

import "math"

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func getMinimumDifference(root *TreeNode) int {
	arr := []int{}
	var f func(root *TreeNode)
	f = func(root *TreeNode) {
		if root == nil {
			return
		}
		f(root.Left)
		arr = append(arr, root.Val)
		f(root.Right)
	}
	f(root)
	ans := math.MaxInt
	for i := 1; i < len(arr); i++ {
		ans = min(ans, arr[i]-arr[i-1])
	}
	return ans
}
