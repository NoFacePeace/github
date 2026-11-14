package str

func longestPalindrome(s string) string {
	n := len(s)
	if n == 0 {
		return ""
	}
	dp := make([][]bool, n)
	for k := range dp {
		dp[k] = make([]bool, n)
	}
	ans := ""
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			if j == i {
				dp[i][j] = true
				continue
			}
			if j == i+1 {
				if s[j] == s[i] {
					dp[i][j] = true
					if j-i+1 > len(ans) {
						ans = s[i : j+1]
					}
				}
				continue
			}
			if dp[i+1][j-1] && s[i] == s[j] {
				dp[i][j] = true
				if j-i+1 > len(ans) {
					ans = s[i : j+1]
				}
			}
		}
	}
	return ans
}
