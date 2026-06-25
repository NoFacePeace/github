package str

func numberOfSpecialChars(word string) int {
	n := len(word)
	m := map[byte]int{}
	for i := 0; i < n; i++ {
		b := word[i]
		if b >= 'a' && b <= 'z' {
			m[b] = i
			continue
		}
		if _, ok := m[b]; !ok {
			m[b] = i
		}
	}
	ans := 0
	for i := 0; i <= 26; i++ {
		b := byte('a' + i)
		if _, ok := m[b]; !ok {
			continue
		}
		o := 'A' + b - 'a'
		if _, ok := m[o]; !ok {
			continue
		}
		if m[b] < m[o] {
			ans++
		}
	}
	return ans
}
