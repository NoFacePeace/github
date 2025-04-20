package bitwise

import "strconv"

func countSymmetricIntegers(low int, high int) int {
	cnt := 0
	for i := low; i <= high; i++ {
		str := strconv.Itoa(i)
		n := len(str)
		if n%2 != 0 {
			continue
		}
		left := 0
		right := 0
		for i := 0; i < n/2; i++ {
			left += int(str[i] - '0')
			right += int(str[n-i-1] - '0')
		}
		if left == right {
			cnt++
		}
	}
	return cnt
}
