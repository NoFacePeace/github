package str

func makeFancyString(s string) string {
	n := len(s)
	if n == 0 {
		return ""
	}
	ans := []byte{}
	cnt := 0
	for i := 0; i < n; i++ {
		if i == 0 {
			ans = append(ans, s[i])
			cnt = 1
			continue
		}
		if s[i] == s[i-1] {
			cnt++
		} else {
			cnt = 1
		}
		if cnt < 3 {
			ans = append(ans, s[i])
		}
	}
	return string(ans)
}
