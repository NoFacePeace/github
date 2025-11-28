package hash

func countPalindromicSubsequence(s string) int {
	n := len(s)
	ans := 0
	visited := map[byte]bool{}
	for i := 0; i < n; i++ {
		c := s[i]
		if visited[c] {
			continue
		}
		visited[c] = true
		j := n - 1
		for j > i {
			if s[j] == c {
				break
			}
			j--
		}
		if j == i {
			continue
		}
		m := map[byte]bool{}
		for k := i + 1; k < j; k++ {
			m[s[k]] = true
		}
		ans += len(m)
	}
	return ans
}
