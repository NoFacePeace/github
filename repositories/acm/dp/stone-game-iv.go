package dp

func winnerSquareGame(n int) bool {
	var dfs func(n int) bool
	dp := map[int]bool{}
	dfs = func(n int) bool {
		if n == 0 {
			return false
		}
		if _, ok := dp[n]; ok {
			return dp[n]
		}
		win := false
		for i := 1; i <= n; i++ {
			if i*i == n {
				win = true
				break
			}
			if i*i > n {
				break
			}
			if !dfs(n - i*i) {
				win = true
				break
			}
		}
		dp[n] = win
		return win
	}
	return dfs(n)
}
