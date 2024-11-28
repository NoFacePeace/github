package pointer

import "math"

func judgeSquareSum(c int) bool {
	left := 0
	right := int(math.Sqrt(float64(c)))
	for left <= right {
		dist := left*left + right*right - c
		if dist == 0 {
			return true
		}
		if dist > 0 {
			right--
		} else {
			left++
		}
	}
	return false
}
