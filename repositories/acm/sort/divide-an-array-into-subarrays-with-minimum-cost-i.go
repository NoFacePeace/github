package sort

import "sort"

func minimumCost(nums []int) int {
	ans := nums[0]
	nums = nums[1:]
	sort.Ints(nums)
	for i := 0; i < 2; i++ {
		ans += nums[i]
	}
	return ans
}
