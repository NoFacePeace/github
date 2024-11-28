package math

import "math"

func smallestRangeI(nums []int, k int) int {
	mx := math.MinInt
	mn := math.MaxInt
	for _, v := range nums {
		mx = max(mx, v)
		mn = min(mn, v)
	}
	if mx-k <= mn+k {
		return 0
	}
	return mx - k - mn - k
}
