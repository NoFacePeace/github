package dp

func superEggDrop(k int, n int) int {
	t := n
	dp := make([][]int, t+1)
	for i := 0; i <= t; i++ {
		dp[i] = make([]int, k+1)
	}
	var f func(t, k int) int
	f = func(t, k int) int {
		if dp[t][k] != 0 {
			return dp[t][k]
		}
		if t == 1 {
			dp[t][k] = 1
			return 1
		}
		if k == 1 {
			dp[t][k] = t
			return t
		}
		cnt := 1 + f(t-1, k-1) + f(t-1, k)
		dp[t][k] = cnt
		return cnt
	}
	f(t, k)
	ans := n
	for i := 1; i <= t; i++ {
		if dp[i][k] >= n {
			ans = i
			break
		}
	}
	return ans
}
