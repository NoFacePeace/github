package str

import "strconv"

func getNoZeroIntegers(n int) []int {
	check := func(num int) bool {
		str := strconv.Itoa(num)
		for i := 0; i < len(str); i++ {
			if str[i] == '0' {
				return false
			}
		}
		return true
	}
	for i := 1; i < n; i++ {
		if check(i) && check(n-i) {
			return []int{i, n - i}
		}
	}
	return []int{}
}
