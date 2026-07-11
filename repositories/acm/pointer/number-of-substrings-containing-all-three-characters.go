package pointer

func numberOfSubstrings(s string) int {
	ans := 0
	m := map[byte]int{}
	l, r := 0, 0
	n := len(s)
	for l < n {
		if m['a'] != 0 && m['b'] != 0 && m['c'] != 0 {
			ans += n - r + 1
			m[s[l]]--
			l++
			continue
		}
		if r == n {
			break
		}
		m[s[r]]++
		r++
	}
	return ans
}
