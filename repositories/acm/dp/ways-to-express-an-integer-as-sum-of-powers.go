package dp

func numberOfWays(n int, x int) int {
	dp := make([]int, n+1)
	dp[0] = 1
	mod := int(1e9) + 7
	for i := 1; i <= n; i++ {
		num := i
		for j := 1; j < x; j++ {
			num *= i
		}
		if num > n {
			break
		}
		for j := n; j >= num; j-- {
			dp[j] = (dp[j] + dp[j-num]) % mod
		}
	}
	return dp[n]
}
