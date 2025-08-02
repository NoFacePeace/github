package hash

import "math"

func maxSum(nums []int) int {
	m := map[int]bool{}
	ans := 0
	mn := math.MinInt
	for _, v := range nums {
		if m[v] {
			continue
		}
		m[v] = true
		if v > 0 {
			ans += v
		} else {
			mn = max(mn, v)
		}
	}
	if ans > 0 {
		return ans
	}
	return mn
}
