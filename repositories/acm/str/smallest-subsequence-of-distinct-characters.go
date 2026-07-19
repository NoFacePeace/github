package str

func smallestSubsequence(s string) string {
	m := map[byte]int{}
	n := len(s)
	for i := 0; i < n; i++ {
		c := s[i]
		m[c]++
	}
	stack := []byte{}
	exist := map[byte]bool{}
	for i := 0; i < n; i++ {
		c := s[i]
		m[c]--
		if exist[c] {
			continue
		}
		for len(stack) > 0 {
			l := len(stack)
			last := stack[l-1]
			if last <= c {
				break
			}
			if m[last] <= 0 {
				break
			}
			stack = stack[:l-1]
			exist[last] = false
		}
		if exist[c] {
			continue
		}
		stack = append(stack, c)
		exist[c] = true
	}
	return string(stack)
}
