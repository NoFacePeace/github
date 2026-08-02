package dp

func stoneGame(piles []int) bool {
	n := len(piles)
	dp := make([][]int, n)
	for i := range dp {
		dp[i] = make([]int, n)
	}
	var dfs func(l, r int) int
	dfs = func(l, r int) int {
		if l == r {
			return piles[l]
		}
		if dp[l][r] != 0 {
			return dp[l][r]
		}
		diff1 := dfs(l+1, r)
		diff2 := dfs(l, r-1)
		if diff1+piles[l] > diff2+piles[r] {
			dp[l][r] = diff1
		} else {
			dp[l][r] = diff2
		}

		return dp[l][r]
	}
	diff := dfs(0, n-1)
	return diff > 0
}
