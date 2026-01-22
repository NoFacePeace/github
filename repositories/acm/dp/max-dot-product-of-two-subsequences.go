package dp

import "math"

func maxDotProduct(nums1 []int, nums2 []int) int {
	m := len(nums1)
	n := len(nums2)
	dp := make([][]int, m)
	for i := 0; i < m; i++ {
		dp[i] = make([]int, n)
		for j := 0; j < n; j++ {
			dp[i][j] = math.MinInt
		}
	}
	var f func(nums1 []int, nums2 []int) int
	f = func(nums1 []int, nums2 []int) int {
		if len(nums1) == 0 {
			return math.MinInt
		}
		if len(nums2) == 0 {
			return math.MinInt
		}
		if dp[m-len(nums1)][n-len(nums2)] != math.MinInt {
			return dp[m-len(nums1)][n-len(nums2)]
		}
		num1 := nums1[0]
		num2 := nums2[0]
		mx := num1 * num2
		for k, v := range nums2 {
			val := v * num1
			ret := f(nums1[1:], nums2[k+1:])
			if ret != math.MinInt {
				val += ret
			}
			mx = max(mx, val)
		}
		mx = max(mx, f(nums1[1:], nums2))
		for k, v := range nums1 {
			val := v * num2
			ret := f(nums1[k+1:], nums2[1:])
			if ret != math.MinInt {
				val += ret
			}
			mx = max(mx, val)
		}
		mx = max(mx, f(nums1, nums2[1:]))
		dp[m-len(nums1)][n-len(nums2)] = mx
		return mx
	}
	f(nums1, nums2)
	ans := math.MinInt
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			ans = max(ans, dp[i][j])
		}
	}
	return ans
}
