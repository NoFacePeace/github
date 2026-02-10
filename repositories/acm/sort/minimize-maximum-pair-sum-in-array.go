package sort

import "sort"

func minPairSum(nums []int) int {
	sort.Ints(nums)
	n := len(nums)
	ans := nums[0] + nums[n-1]
	for i := 1; i < n/2; i++ {
		ans = max(ans, nums[i]+nums[n-i-1])
	}
	return ans
}
