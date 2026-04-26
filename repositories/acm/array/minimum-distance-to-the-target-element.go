package array

import "math"

func getMinDistance(nums []int, target int, start int) int {
	abs := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	ans := math.MaxInt
	for k, v := range nums {
		if v == target {
			ans = min(ans, abs(k, start))
		}
	}
	return ans
}
