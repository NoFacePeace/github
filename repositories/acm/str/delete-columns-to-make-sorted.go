package str

func minDeletionSize(strs []string) int {
	n := len(strs)
	if n == 0 {
		return 0
	}
	m := map[int]bool{}
	for i := 0; i < len(strs[0]); i++ {
		for j := 0; j < n; j++ {
			if j == 0 {
				continue
			}
			if strs[j][i] < strs[j-1][i] {
				m[i] = true
			}
		}
	}
	return len(m)
}
