package tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func averageOfLevels(root *TreeNode) []float64 {
	q := []*TreeNode{root}
	arr := []float64{}
	for len(q) > 0 {
		tmp := q
		q = nil
		sum := 0
		cnt := 0
		for _, v := range tmp {
			if v == nil {
				continue
			}
			sum += v.Val
			cnt++
			q = append(q, v.Left, v.Right)
		}
		if cnt == 0 {
			continue
		}
		arr = append(arr, float64(sum)/float64(cnt))
	}
	return arr
}
