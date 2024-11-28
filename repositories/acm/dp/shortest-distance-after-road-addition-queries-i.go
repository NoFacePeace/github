package dp

func shortestDistanceAfterQueries(n int, queries [][]int) []int {
	prev := make([][]int, n)
	dp := make([]int, n)
	for i := 1; i < n; i++ {
		prev[i] = append(prev[i], i-1)
		dp[i] = i
	}
	ans := []int{}
	for _, query := range queries {
		u, v := query[0], query[1]
		prev[v] = append(prev[v], u)
		for ; v < n; v++ {
			for _, u := range prev[v] {
				dp[v] = min(dp[v], dp[u]+1)
			}
		}
		ans = append(ans, dp[n-1])
	}
	return ans
}
