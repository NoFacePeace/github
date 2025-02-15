package dp

// https://leetcode.cn/problems/unique-paths-ii/

func uniquePathsWithObstacles(obstacleGrid [][]int) int {
	m := len(obstacleGrid)
	if m == 0 {
		return 0
	}
	n := len(obstacleGrid[0])
	if n == 0 {
		return 0
	}
	dp := make([]int, n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if obstacleGrid[i][j] == 1 {
				dp[j] = 0
				continue
			}
			if i == 0 && j == 0 {
				dp[j] = 1
				continue
			}
			if i == 0 {
				dp[j] = dp[j-1]
				continue
			}
			if j == 0 {
				dp[j] = dp[j]
				continue
			}
			dp[j] = dp[j] + dp[j-1]
		}
	}
	return dp[n-1]
}
