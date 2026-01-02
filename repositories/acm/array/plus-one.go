package array

func plusOne(digits []int) []int {
	digits = append([]int{0}, digits...)
	n := len(digits)
	add := 1
	for i := n - 1; i >= 0; i-- {
		if add == 0 {
			break
		}
		digits[i] += add
		if digits[i] == 10 {
			digits[i] = 0
			add = 1
		} else {
			add = 0
		}
	}
	if digits[0] == 0 {
		digits = digits[1:]
	}
	return digits
}
