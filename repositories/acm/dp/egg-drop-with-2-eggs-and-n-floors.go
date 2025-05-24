package dp

func twoEggDrop(n int) int {
	m := map[int]int{}
	var f func(n int) int
	f = func(n int) int {
		if val, ok := m[n]; ok {
			return val
		}
		if n == 0 {
			return 0
		}
		if n == 1 {
			return 1
		}
		ans := n
		for i := 1; i < n; i++ {
			cnt := maxSlice(i, f(n-i)+1)
			ans = min(ans, cnt)
		}
		m[n] = ans
		return ans
	}
	return f(n)
}
