package dp

// https://leetcode.cn/problems/find-minimum-time-to-reach-last-room-i/?envType=daily-question&envId=2025-05-07

import "math"

func minTimeToReach(moveTime [][]int) int {
	n := len(moveTime)
	if n == 0 {
		return 0
	}
	m := len(moveTime[0])
	if m == 0 {
		return 0
	}
	dp := make([][]int, n)
	for i := 0; i < n; i++ {
		dp[i] = make([]int, m)
		for j := 0; j < m; j++ {
			dp[i][j] = math.MaxInt
		}
	}
	var dfs func(i, j int)
	dfs = func(i, j int) {
		if i-1 >= 0 {
			t := max(dp[i][j], moveTime[i-1][j])
			if t+1 < dp[i-1][j] {
				dp[i-1][j] = t + 1
				dfs(i-1, j)
			}
		}
		if i+1 < n {
			t := max(dp[i][j], moveTime[i+1][j])
			if t+1 < dp[i+1][j] {
				dp[i+1][j] = t + 1
				dfs(i+1, j)
			}
		}
		if j-1 >= 0 {
			t := max(dp[i][j], moveTime[i][j-1])
			if t+1 < dp[i][j-1] {
				dp[i][j-1] = t + 1
				dfs(i, j-1)
			}
		}
		if j+1 < m {
			t := max(dp[i][j], moveTime[i][j+1])
			if t+1 < dp[i][j+1] {
				dp[i][j+1] = t + 1
				dfs(i, j+1)
			}
		}
	}
	dp[0][0] = 0
	dfs(0, 0)
	return dp[n-1][m-1]
}
