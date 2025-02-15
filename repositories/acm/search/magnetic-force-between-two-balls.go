package search

import "sort"

func maxDistance(position []int, m int) int {
	sort.Ints(position)
	left, right := 1, position[len(position)-1]-position[0]
	ans := -1
	check := func(x int) bool {
		pre, cnt := position[0], 1
		for i := 1; i < len(position); i++ {
			if position[i]-pre >= x {
				pre = position[i]
				cnt++
			}

		}
		return cnt >= m
	}
	for left <= right {
		mid := (left + right) / 2
		if check(mid) {
			ans = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return ans
}
