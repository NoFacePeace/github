package sort

import (
	"sort"
)

func minAnagramLength(s string) int {
	n := len(s)
	if n == 0 {
		return 0
	}
	check := func(arr []string) bool {
		s := []byte(arr[0])
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		for i := 1; i < len(arr); i++ {
			tmp := []byte(arr[i])
			sort.Slice(tmp, func(i, j int) bool {
				return tmp[i] < tmp[j]
			})
			if string(s) != string(tmp) {
				return false
			}
		}
		return true
	}
	split := func(s string, n int) []string {
		arr := []string{}
		for i := 0; i < len(s); i += len(s) / n {
			arr = append(arr, s[i:i+len(s)/n])
		}
		return arr
	}
	for i := n; i > 0; i-- {
		if n/i*i != n {
			continue
		}
		arr := split(s, i)
		if check(arr) {
			return n / i
		}
	}
	return n
}
