package math

func sumAndMultiply(n int) int64 {
	sum := 0
	num := 0
	power := 1
	for n != 0 {
		mod := n % 10
		n /= 10
		if mod == 0 {
			continue
		}
		sum += mod
		num += power * mod
		power *= 10
	}
	return int64(num * sum)
}
