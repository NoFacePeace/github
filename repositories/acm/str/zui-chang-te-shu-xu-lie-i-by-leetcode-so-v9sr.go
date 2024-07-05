package str

func findLUSlength(a string, b string) int {
	if a == b {
		return -1
	}
	l := len(a)
	if l < len(b) {
		l = len(b)
	}
	return l
}
