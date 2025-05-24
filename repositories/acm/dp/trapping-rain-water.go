package dp

func trap(height []int) int {
	n := len(height)
	if n == 0 {
		return 0
	}
	ans := 0
	lm := make([]int, n)
	lm[0] = height[0]
	for i := 1; i < n; i++ {
		lm[i] = maxSlice(lm[i-1], height[i])
	}
	rm := make([]int, n)
	rm[n-1] = height[n-1]
	for i := n - 2; i >= 0; i-- {
		rm[i] = maxSlice(rm[i+1], height[i])
	}
	for i, h := range height {
		ans += min(lm[i], rm[i]) - h
	}
	return ans
}
