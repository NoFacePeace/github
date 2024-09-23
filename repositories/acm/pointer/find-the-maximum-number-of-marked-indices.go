package pointer

import "sort"

func maxNumOfMarkedIndices(nums []int) int {
	sort.Ints(nums)
	ans := 0
	n := len(nums)
	l, r := 0, n/2
	for l < n/2 && r < n {
		if nums[l]*2 <= nums[r] {
			ans += 2
			l++
			r++
			continue
		}
		r++
	}
	return ans
}
