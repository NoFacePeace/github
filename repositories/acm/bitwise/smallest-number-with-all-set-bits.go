package bitwise

func smallestNumber(n int) int {
	num := 2
	for num-1 < n {
		num = num << 1
	}
	return num - 1
}
