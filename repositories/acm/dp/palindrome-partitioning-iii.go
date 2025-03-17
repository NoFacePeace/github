package dp

func palindromePartition(s string, k int) int {
	n := len(s)
	g := make([][]int, n)
	for i := range g {
		g[i] = make([]int, n)
	}
	for i := n - 1; i >= 0; i-- {
		for j := i + 1; j < n; j++ {
			if s[i] == s[j] {
				g[i][j] = g[i+1][j-1]
			} else {
				g[i][j] = g[i+1][j-1] + 1
			}
		}
	}
	f := make([]int, n)
	for i := 0; i < k; i++ {
		for j := n - 1; j >= 0; j-- {
			if i == 0 {
				f[j] = g[0][j]
				continue
			}
			f[j] = j
			for x := i - 1; x < j; x++ {
				f[j] = min(f[j], f[x]+g[x+1][j])
			}
		}
	}
	return f[n-1]
}
