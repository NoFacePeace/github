package search

func maxJumps(arr []int, d int) int {
	n := len(arr)
	dp := make([]int, n)
	ans := 0
	var dfs func(idx int) int
	dfs = func(idx int) int {
		if dp[idx] > 0 {
			return dp[idx]
		}
		dp[idx] = 1
		for i := idx + 1; i <= min(n-1, idx+d); i++ {
			if arr[i] >= arr[idx] {
				break
			}
			dp[idx] = max(dp[idx], dfs(i)+1)
		}
		for i := idx - 1; i >= max(0, idx-d); i-- {
			if arr[i] >= arr[idx] {
				break
			}
			dp[idx] = max(dp[idx], dfs(i)+1)
		}
		ans = max(ans, dp[idx])
		return dp[idx]
	}
	for i := 0; i < n; i++ {
		if dp[i] > 0 {
			continue
		}
		dfs(i)
	}
	return ans
}
