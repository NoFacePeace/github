package bitwise

func isPowerOfFour(n int) bool {
	num := 1
	for num <= n {
		if num == n {
			return true
		}
		num *= 4
	}
	return false
}
