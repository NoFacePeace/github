package bitwise

func maxProduct(n int) int {
	m := map[int]int{}
	for n != 0 {
		bit := n % 10
		n /= 10
		m[bit]++
	}
	last := -1
	for i := 9; i >= 0; i-- {
		cnt, ok := m[i]
		if !ok {
			continue
		}
		if last != -1 {
			return last * i
		}
		if cnt > 1 {
			return i * i
		}
		last = i
	}
	return 0
}
