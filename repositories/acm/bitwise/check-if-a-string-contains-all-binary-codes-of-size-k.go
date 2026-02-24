package bitwise

func hasAllCodes(s string, k int) bool {
	n := len(s)
	m := map[string]bool{}
	for i := 0; i <= n-k; i++ {
		sub := s[i : i+k]
		m[sub] = true
	}
	cnt := 1
	for i := 0; i < k; i++ {
		cnt *= 2
	}
	if len(m) != cnt {
		return false
	}
	return true
}
