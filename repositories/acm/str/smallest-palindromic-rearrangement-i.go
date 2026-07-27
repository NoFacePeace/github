package str

func smallestPalindrome(s string) string {
	m := map[byte]string{}
	n := len(s)
	for i := 0; i < n; i++ {
		c := s[i]
		m[c] += s[i : i+1]
	}
	prefix := ""
	suffix := ""
	mid := ""
	for i := byte('a'); i <= byte('z'); i++ {
		l := len(m[i])
		if l == 0 {
			continue
		}
		if l%2 == 1 {
			mid = m[i][:1]
			m[i] = m[i][1:]
		}
		prefix += m[i][:l/2]
		suffix = m[i][l/2:] + suffix
	}
	return prefix + mid + suffix
}
