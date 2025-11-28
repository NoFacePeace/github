package math

func nextBeautifulNumber(n int) int {
	cnt := 0
	mn := n
	for n != 0 {
		n /= 10
		cnt++
	}
	cnt += 1
	mx := 0
	for i := 0; i < cnt; i++ {
		mx *= 10
		mx += cnt
	}
	check := func(num int) bool {
		m := map[int]int{}
		for num != 0 {
			mod := num % 10
			m[mod]++
			num /= 10
		}
		for k, v := range m {
			if k != v {
				return false
			}
		}
		return true
	}
	for i := mn + 1; i <= mx; i++ {
		if check(i) {
			return i
		}
	}
	return mx
}
