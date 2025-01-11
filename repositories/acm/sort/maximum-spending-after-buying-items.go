package sort

import "sort"

func maxSpending(values [][]int) int64 {
	nums := []int{}
	for i := 0; i < len(values); i++ {
		for j := 0; j < len(values[i]); j++ {
			nums = append(nums, values[i][j])
		}
	}
	sort.Ints(nums)
	sum := 0
	for i := range nums {
		sum += nums[i] * (i + 1)
	}
	return int64(sum)
}
