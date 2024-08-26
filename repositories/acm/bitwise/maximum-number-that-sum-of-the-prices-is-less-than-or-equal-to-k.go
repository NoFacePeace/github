package bitwise

func findMaximumNumber(k int64, x int) int64 {
	f := func(num int) int {
		cnt := 0
		i := 1
		for {
			bit := 1 << (x*i - 1)
			if bit > num {
				break
			}
			if bit&num != 0 {
				cnt++
			}
			i++
		}
		return cnt
	}
	num := 1
	sum := 0
	for sum <= int(k) {
		val := f(num)
		if sum+val > int(k) {
			num = num - 1
			break
		}
		sum += val
		num = num + 1
	}
	return int64(num)
}
