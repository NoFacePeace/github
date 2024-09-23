package str

func longestContinuousSubstring(s string) int {
	n := len(s)
	if n == 0 {
		return 0
	}
	cnt := 1
	ans := 1
	for i := 1; i < n; i++ {
		if s[i] == s[i-1]+1 {
			cnt++
		} else {
			cnt = 1
		}
		ans = max(ans, cnt)
	}
	return ans
}
