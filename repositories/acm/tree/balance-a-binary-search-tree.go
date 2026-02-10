package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func balanceBST(root *TreeNode) *TreeNode {
	nums := []int{}
	var dfs func(root *TreeNode)
	dfs = func(root *TreeNode) {
		if root == nil {
			return
		}
		dfs(root.Left)
		nums = append(nums, root.Val)
		dfs(root.Right)
	}
	dfs(root)
	var build func(nums []int) *TreeNode
	build = func(nums []int) *TreeNode {
		n := len(nums)
		if n == 0 {
			return nil
		}
		if n == 1 {
			return &TreeNode{
				Val: nums[0],
			}
		}
		mid := n / 2
		root := &TreeNode{
			Val: nums[mid],
		}
		root.Left = build(nums[:mid])
		root.Right = build(nums[mid+1:])
		return root
	}
	return build(nums)
}
