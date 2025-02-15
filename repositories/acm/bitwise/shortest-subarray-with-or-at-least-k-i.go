package bitwise

import "math"

func minimumSubarrayLength(nums []int, k int) int {
	n := len(nums)
	bits := make([]int, 30)
	res := math.MaxInt32
	for l, r := 0, 0; r < n; r++ {
		for i := 0; i < 30; i++ {
			bits[i] += (nums[r] >> i) & 1
		}
		for l <= r && calc(bits) >= k {
			res = min(res, r-l+1)
			for i := 0; i < 30; i++ {
				bits[i] -= (nums[l] >> i) & 1
			}
			l++
		}
	}
	if res == math.MaxInt32 {
		return -1
	}
	return res
}

func calc(bits []int) int {
	ans := 0
	for i := 0; i < len(bits); i++ {
		if bits[i] > 0 {
			ans |= 1 << i
		}
	}
	return ans
}
