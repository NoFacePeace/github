package dp

func isInterleave(s1 string, s2 string, s3 string) bool {
	m := map[int]bool{}
	n := len(s1)
	var f func(int, int, int) bool
	f = func(a, b, c int) bool {
		if _, ok := m[a+b*n]; ok {
			return m[a*n+b]
		}
		if a == len(s1) && b == len(s2) && c == len(s3) {
			return true
		}
		if c == len(s3) {
			return false
		}
		if a < len(s1) && s1[a] == s3[c] && f(a+1, b, c+1) {
			m[a+n*b] = true
			return true
		}
		if b < len(s2) && s2[b] == s3[c] && f(a, b+1, c+1) {
			m[a+n*b] = true
			return true
		}
		m[a+n*b] = false
		return false
	}
	return f(0, 0, 0)
}
