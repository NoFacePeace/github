package bitwise

func smallestNumberI(n int, t int) int {
	for {
		tmp := n
		num := 1
		for tmp != 0 {
			bit := tmp % 10
			tmp /= 10
			num *= bit
		}
		if num%t == 0 {
			return n
		}
		n++
	}
	return 0
}
