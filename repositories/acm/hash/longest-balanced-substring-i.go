package hash

func longestBalancedI(s string) int {
	n := len(s)
	ans := 0
	check := func(m map[byte]int) bool {
		cnt := -1
		for _, v := range m {
			if cnt == -1 {
				cnt = v
				continue
			}
			if cnt != v {
				return false
			}
		}
		return true
	}
	for i := 0; i < n; i++ {
		m := map[byte]int{}
		for j := i; j < n; j++ {
			c := s[j]
			m[c]++
			if check(m) {
				ans = max(ans, j-i+1)
			}
		}
	}
	return ans
}
