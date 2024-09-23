package dp

func maximalSquare(matrix [][]byte) int {
	m := len(matrix)
	if m == 0 {
		return 0
	}
	n := len(matrix[0])
	if n == 0 {
		return 0
	}
	mx := 0
	dp := make([][]int, m)
	for i := 0; i < m; i++ {
		dp[i] = make([]int, n)
		for j := 0; j < n; j++ {
			if matrix[i][j] == '0' {
				continue
			}
			dp[i][j] = 1
			mx = max(dp[i][j], mx)
			if i == 0 {
				continue
			}
			if j == 0 {
				continue
			}
			mn := min(dp[i][j-1], dp[i-1][j]) + 1
			dp[i][j] = min(mn, dp[i-1][j-1]+1)
			mx = max(dp[i][j], mx)
		}
	}
	return mx * mx
}
