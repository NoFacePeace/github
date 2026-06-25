package str

func rotateString(s string, goal string) bool {
	n := len(s)
	for i := 0; i < n; i++ {
		s = s[1:] + s[:1]
		if s == goal {
			return true
		}
	}
	return false
}
