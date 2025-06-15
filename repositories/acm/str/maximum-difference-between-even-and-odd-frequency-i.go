package str

import "math"

func maxDifference(s string) int {
	m := map[rune]int{}
	for _, v := range s {
		m[v]++
	}
	even := math.MaxInt
	odd := 0
	for _, v := range m {
		if v%2 == 0 {
			even = min(even, v)
		} else {
			odd = max(odd, v)
		}
	}
	return odd - even
}
