package hash

func reorderedPowerOf2(n int) bool {
	m := map[int]int{}
	for n > 0 {
		num := n % 10
		m[num]++
		n /= 10
	}
	mx := int(1e9)
	num := 1
	for num <= mx {
		m1 := map[int]int{}
		tmp := num
		for tmp > 0 {
			m1[tmp%10]++
			tmp /= 10
		}
		ok := true
		for i := 0; i < 9; i++ {
			if m[i] != m1[i] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
		num *= 2
	}
	return false
}
