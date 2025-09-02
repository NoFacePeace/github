package math

func isPowerOfThree(n int) bool {
	num := 1
	for num <= n {
		if num == n {
			return true
		}
		num *= 3
	}
	return false
}
