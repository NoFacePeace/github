package str

import "math"

func maxNumberOfBalloons(text string) int {
	m := map[byte]int{
		'b': 0,
		'a': 0,
		'l': 0,
		'o': 0,
		'n': 0,
	}
	n := len(text)
	for i := 0; i < n; i++ {
		c := text[i]
		if _, ok := m[c]; !ok {
			continue
		}
		m[c]++
	}
	ans := math.MaxInt
	for k, v := range m {
		if k == 'l' || k == 'o' {
			ans = min(ans, v/2)
			continue
		}
		ans = min(ans, v)
	}
	return ans
}
