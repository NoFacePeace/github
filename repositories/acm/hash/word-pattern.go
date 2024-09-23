package hash

import "strings"

func wordPattern(pattern string, s string) bool {
	arr := strings.Split(s, " ")
	m := map[string]byte{}
	n := len(arr)
	if n != len(pattern) {
		return false
	}
	code := []byte{}
	used := map[byte]bool{}
	for i := 0; i < n; i++ {
		str := arr[i]
		if _, ok := m[str]; !ok {
			if used[pattern[i]] {
				return false
			}
			m[str] = pattern[i]
			used[pattern[i]] = true
		}
		code = append(code, m[str])
	}
	return string(code) == pattern
}
