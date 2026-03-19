package str

func checkOnesSegment(s string) bool {
	n := len(s)
	cnt := 0
	for i := n - 1; i >= 0; i-- {
		if s[i] == '1' {
			cnt++
			continue
		}
		if cnt != 0 {
			return false
		}
	}
	return true
}
