package dp

func longestValidParentheses(s string) int {
	n := len(s)
	dp := make([]int, n+1)
	ans := 0
	for i := 0; i < n; i++ {
		if i == 0 {
			continue
		}
		if s[i-1] == '(' && s[i-1] != s[i] {
			dp[i+1] = max(dp[i+1], dp[i-1]+2)
		}
		left := i - dp[i] - 1
		if dp[i] != 0 && left >= 0 && s[left] == '(' && s[left] != s[i] {
			dp[i+1] = max(dp[i+1], dp[i]+2+dp[left])
		}
		ans = max(ans, dp[i+1])
	}
	return ans
}
