package dp

func mincostTickets(days []int, costs []int) int {
	trip := map[int]bool{}
	for _, day := range days {
		trip[day] = true
	}
	dp := make([]int, 366)
	for i := 1; i < 366; i++ {
		if !trip[i] {
			dp[i] = dp[i-1]
			continue
		}
		dp[i] = costs[0] + dp[i-1]
		if i-7 >= 0 {
			dp[i] = min(dp[i], dp[i-7]+costs[1])
		} else {
			dp[i] = min(dp[i], costs[1])
		}
		if i-30 >= 0 {
			dp[i] = min(dp[i], dp[i-30]+costs[2])
		} else {
			dp[i] = min(dp[i], costs[2])
		}
	}
	return dp[365]
}
