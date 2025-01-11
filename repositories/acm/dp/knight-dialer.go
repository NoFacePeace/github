package dp

func knightDialer(n int) int {
	mod := int(1e9) + 7
	dp := make([][]int, 10)
	for i := 0; i < 10; i++ {
		dp[i] = make([]int, n)
		dp[i][0] = 1
	}
	for i := 1; i < n; i++ {
		dp[0][i] = (dp[4][i-1] + dp[6][i-1]) % mod
		dp[1][i] = (dp[8][i-1] + dp[6][i-1]) % mod
		dp[2][i] = (dp[7][i-1] + dp[9][i-1]) % mod
		dp[3][i] = (dp[4][i-1] + dp[8][i-1]) % mod
		dp[4][i] = (dp[3][i-1] + dp[9][i-1] + dp[0][i-1]) % mod
		dp[6][i] = (dp[1][i-1] + dp[7][i-1] + dp[0][i-1]) % mod
		dp[7][i] = (dp[2][i-1] + dp[6][i-1]) % mod
		dp[8][i] = (dp[1][i-1] + dp[3][i-1]) % mod
		dp[9][i] = (dp[2][i-1] + dp[4][i-1]) % mod
	}
	ans := 0
	for i := 0; i < 10; i++ {
		ans += dp[i][n-1]
		ans %= mod
	}
	return ans
}
