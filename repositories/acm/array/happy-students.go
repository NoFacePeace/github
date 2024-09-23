package array

import "sort"

func countWays(nums []int) int {
	sort.Ints(nums)
	n := len(nums)
	if n == 0 {
		return 1
	}
	cnt := 0
	if nums[0] != 0 {
		cnt++
	}
	for i := n - 1; i >= 0; i-- {
		if nums[i] >= i+1 {
			continue
		}
		if i < n-1 && nums[i] == nums[i+1] {
			continue
		}
		if i < n-1 && i+1 >= nums[i+1] {
			continue
		}
		cnt++
	}
	return cnt
}
