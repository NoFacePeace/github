package dp

// https://leetcode.cn/problems/solving-questions-with-brainpower/?envType=daily-question&envId=2025-04-01

func mostPoints(questions [][]int) int64 {
	n := len(questions)
	dp := make([]int, n+1)
	for i := n - 1; i >= 0; i-- {
		dp[i] = maxSlice(dp[i+1], questions[i][0]+dp[min(n, i+questions[i][1]+1)])
	}
	return int64(dp[0])
}
