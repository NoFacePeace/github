package hash

func minimizedStringLength(s string) int {
	m := map[byte]bool{}
	n := len(s)
	ans := []byte{}
	for i := 0; i < n; i++ {
		c := s[i]
		if _, ok := m[c]; ok {
			continue
		}
		m[c] = true
		ans = append(ans, c)
	}
	return len(ans)
}
