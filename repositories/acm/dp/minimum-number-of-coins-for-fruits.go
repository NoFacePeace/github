package dp

func minimumCoins(prices []int) int {
	n := len(prices)
	if n == 0 {
		return 0
	}
	dp := make([]int, n+1)
	for i := 1; i <= n; i++ {
		for j := i; j <= min(n, i*2); j++ {
			if dp[j] == 0 {
				dp[j] = dp[i-1] + prices[i-1]
				continue
			}
			dp[j] = min(dp[j], dp[i-1]+prices[i-1])
		}
	}
	return dp[n]
}
