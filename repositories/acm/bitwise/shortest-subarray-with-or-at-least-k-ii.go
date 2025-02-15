package bitwise

import "math"

func minimumSubarrayLengthII(nums []int, k int) int {
	l, r := 0, 0
	n := len(nums)
	bits := make([]int, 30)
	update := func(num int, val int) {
		i := 0
		for num > 0 {
			v := num & 1
			num = num >> 1
			if v == 1 {
				bits[i] += val
			}
			i++
		}
	}
	calc := func() int {
		num := 0
		for i := 0; i < len(bits); i++ {
			if bits[i] > 0 {
				num += 1 << i
			}
		}
		return num
	}
	ans := math.MaxInt
	for r < n {
		update(nums[r], 1)
		r++
		for l < r && calc() >= k {
			ans = min(ans, r-l)
			update(nums[l], -1)
			l++
		}
	}
	if ans == math.MaxInt {
		return -1
	}
	return ans
}
