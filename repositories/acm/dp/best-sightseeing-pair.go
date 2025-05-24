package dp

func maxScoreSightseeingPair(values []int) int {
	n := len(values)
	if n < 2 {
		return 0
	}
	dp := make([]int, n)
	dp[1] = values[0] + values[1] - 1
	ans := dp[1]
	for i := 2; i < n; i++ {
		dp[i] = maxSlice(values[i]+values[i-1]-1, dp[i-1]-values[i-1]+values[i]-1)
		ans = maxSlice(ans, dp[i])
	}
	return ans
}
