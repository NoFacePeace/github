package bitwise

func isSubstringPresent(s string) bool {
	h := make([]int, 26)
	for i := 0; i+1 < len(s); i++ {
		x, y := s[i]-'a', s[i+1]-'a'
		h[x] |= (1 << y)
		if (h[y]>>x)&1 != 0 {
			return true
		}
	}
	return false
}
