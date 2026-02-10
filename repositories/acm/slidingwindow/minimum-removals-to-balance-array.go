package slidingwindow

import "sort"

func minRemoval(nums []int, k int) int {
	n := len(nums)
	sort.Ints(nums)
	if n == 0 {
		return 0
	}
	l, r := 0, 0
	ans := n - 1
	for r < n {
		if l == r {
			r++
			continue
		}
		if nums[l]*k < nums[r] {
			l++
			continue
		}
		ans = min(ans, n-r+l-1)
		r++
	}
	return ans
}
