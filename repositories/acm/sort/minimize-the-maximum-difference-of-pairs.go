package sort

import "sort"

func minimizeMax(nums []int, p int) int {
	sort.Ints(nums)
	n := len(nums)
	check := func(mx int) bool {
		cnt := 0
		for i := 0; i < n-1; i++ {
			if nums[i+1]-nums[i] <= mx {
				cnt++
				i++
			}
		}
		return cnt >= p
	}
	left, right := 0, nums[n-1]-nums[0]
	for left < right {
		mid := (left + right) / 2
		if check(mid) {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}
