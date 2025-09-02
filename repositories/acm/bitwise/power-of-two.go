package bitwise

func isPowerOfTwo(n int) bool {
	val := 1
	for n >= val {
		if n == val {
			return true
		}
		val = val << 1
	}
	return false
}
