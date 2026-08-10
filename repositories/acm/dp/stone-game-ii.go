package dp

import "math"

func stoneGameII(piles []int) int {
	n := len(piles)
	if n == 1 {
		return piles[0]
	}
	dp := make([][]int, n)
	for i := range dp {
		dp[i] = make([]int, n)
		for j := range dp[i] {
			dp[i][j] = math.MinInt
		}
	}
	var dfs func(i, m int) int
	dfs = func(i, m int) int {
		cnt := 0
		mx := math.MinInt
		if i == n {
			return 0
		}
		if m >= n {
			m = n - 1
		}
		if dp[i][m] != math.MinInt {
			return dp[i][m]
		}
		for j := 1; j <= 2*m; j++ {
			if i+j-1 >= n {
				break
			}
			cnt += piles[i+j-1]
			val := dfs(i+j, max(m, j))
			mx = max(mx, cnt-val)
		}
		dp[i][m] = mx
		return mx
	}
	dfs(0, 1)
	sum := 0
	for _, v := range piles {
		sum += v
	}
	val := math.MinInt
	for _, v := range dp[0] {
		val = max(val, v)
	}
	return (sum + val) / 2
}
