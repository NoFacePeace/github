package array

import "math"

// https://leetcode.cn/problems/find-closest-number-to-zero/description/

func findClosestNumber(nums []int) int {
	abs := func(num int) int {
		if num < 0 {
			return 0 - num
		}
		return num
	}
	dist := math.MaxInt
	ans := 0
	for _, v := range nums {
		d := abs(v)
		if d < dist {
			dist = d
			ans = v
			continue
		}
		if d == dist && v > ans {
			ans = v
			continue
		}
	}
	return ans
}
