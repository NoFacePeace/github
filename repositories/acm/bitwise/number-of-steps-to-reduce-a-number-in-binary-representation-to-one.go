package bitwise

func numSteps(s string) int {
	add := func(s string) string {
		tmp := ""
		n := len(s)
		for i := n - 1; i >= 0; i-- {
			if s[i] == '0' {
				tmp = "1" + tmp
				tmp = s[:i] + tmp
				return tmp
			}
			tmp = "0" + tmp
		}
		tmp = "1" + tmp
		return tmp
	}
	cnt := 0
	for s != "1" {
		n := len(s)
		c := s[n-1]
		cnt++
		if c == '0' {
			s = s[:n-1]
			continue
		}
		s = add(s)
	}
	return cnt
}
