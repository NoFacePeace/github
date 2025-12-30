package dp

func maximumProfit(prices []int, k int) int64 {
	n := len(prices)
	dp := make([][3]int, k+1)
	for j := 1; j <= k; j++ {
		dp[j][1] = -prices[0]
		dp[j][2] = prices[0]
	}
	for i := 0; i < n; i++ {
		for j := k; j > 0; j-- {
			dp[j][0] = max(dp[j][0], max(dp[j][1]+prices[i]), dp[j][2]-prices[i])
			dp[j][1] = max(dp[j][1], dp[j-1][0]-prices[i])
			dp[j][2] = max(dp[j][2], dp[j-1][0]+prices[i])
		}
	}
	return int64(dp[k][0])
}
