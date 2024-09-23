package dp

func coinChange(coins []int, amount int) int {
	dp := map[int]int{}
	for _, v := range coins {
		if v <= amount {
			dp[v] = 1
		}
	}
	dp[0] = 0
	for i := 0; i <= amount; i++ {
		for _, v := range coins {
			if v > amount {
				continue
			}
			if _, ok := dp[i-v]; !ok {
				continue
			}
			if _, ok := dp[i]; !ok {
				dp[i] = dp[i-v] + 1
				continue
			}
			dp[i] = min(dp[i], dp[i-v]+1)
		}
	}
	if _, ok := dp[amount]; ok {
		return dp[amount]
	}
	return -1
}
