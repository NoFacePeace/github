package str

func minOperations(s string) int {
	n := len(s)
	cnt := 0
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			if s[i] != '0' {
				cnt++
			}
		} else {
			if s[i] != '1' {
				cnt++
			}
		}
	}
	ans := cnt
	cnt = 0
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			if s[i] != '1' {
				cnt++
			}
		} else {
			if s[i] != '0' {
				cnt++
			}
		}
	}
	ans = min(ans, cnt)
	return ans
}
