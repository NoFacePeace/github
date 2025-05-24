package array

// https://leetcode.cn/problems/zero-array-transformation-ii/?envType=daily-question&envId=2025-05-21

func minZeroArray(nums []int, queries [][]int) int {
	n := len(nums)
	deltaArray := make([]int, n+1)
	operations := 0
	k := 0
	for i, num := range nums {
		operations += deltaArray[i]
		for k < len(queries) && operations < num {
			left, right, value := queries[k][0], queries[k][1], queries[k][2]
			deltaArray[left] += value
			deltaArray[right+1] -= value
			if left <= i && i <= right {
				operations += value
			}
			k++
		}
		if operations < num {
			return -1
		}
	}
	return k
}
