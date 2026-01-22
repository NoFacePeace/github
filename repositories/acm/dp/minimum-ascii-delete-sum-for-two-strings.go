package dp

func minimumDeleteSum(s1 string, s2 string) int {
	m := len(s1)
	n := len(s2)
	dp := make([][]int, m+1)
	for i := 0; i <= m; i++ {
		dp[i] = make([]int, n+1)
		for j := 0; j <= n; j++ {
			if i == 0 && j == 0 {
				continue
			}
			if i == 0 {
				dp[i][j] = int(s2[j-1]) + dp[i][j-1]
			}
			if j == 0 {
				dp[i][j] = int(s1[i-1]) + dp[i-1][j]
			}
		}
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			dp[i+1][j+1] = min(dp[i+1][j]+int(s2[j]), dp[i][j+1]+int(s1[i]))
			if s1[i] == s2[j] {
				dp[i+1][j+1] = min(dp[i][j], dp[i+1][j+1])
			}
		}
	}
	return dp[m][n]
}
