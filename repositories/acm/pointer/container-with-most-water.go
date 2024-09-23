package pointer

import "math"

func maxArea(height []int) int {
	l := 0
	r := len(height) - 1
	mx := math.MinInt
	for l < r {
		area := (r - l) * min(height[l], height[r])
		if area > mx {
			mx = area
		}
		if height[l] < height[r] {
			l++
		} else {
			r--
		}
	}
	return mx
}
