package sort

import "sort"

func partitionArray(nums []int, k int) int {
	sort.Ints(nums)
	n := len(nums)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	start := nums[0]
	ans := 0
	for i := 1; i < n; i++ {
		if nums[i]-start <= k {
			if i == n-1 {
				ans++
			}
			continue
		}
		start = nums[i]
		ans++
		if i == n-1 {
			ans++
		}
	}
	return ans
}
