package dp

func wordBreak(s string, wordDict []string) bool {
	n := len(s)
	dp := make([]bool, n)
	m := map[string]bool{}
	for _, v := range wordDict {
		m[v] = true
	}
	for i := 0; i < n; i++ {
		word := s[:i+1]
		if m[word] {
			dp[i] = true
			continue
		}
		for j := 0; j < i; j++ {
			if dp[j] && m[s[j+1:i+1]] {
				dp[i] = true
				break
			}
		}
	}
	return dp[n-1]
}
