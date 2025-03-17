package dp

// https://leetcode.cn/problems/palindrome-partitioning-iv/description/
func checkPartitioning(s string) bool {
	n := len(s)
	g := make([][]bool, n)
	for i := range g {
		g[i] = make([]bool, n)
	}
	for i := n - 1; i >= 0; i-- {
		for j := i; j < n; j++ {
			if i == j {
				g[i][j] = true
				continue
			}
			if i+1 == j {
				g[i][j] = s[i] == s[j]
				continue
			}
			g[i][j] = s[i] == s[j] && g[i+1][j-1]
		}
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n-1; j++ {
			if g[0][i] && g[i+1][j] && g[j+1][n-1] {
				return true
			}
		}
	}
	return false
}
