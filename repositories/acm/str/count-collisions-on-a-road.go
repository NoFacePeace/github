package str

func countCollisions(directions string) int {
	n := len(directions)
	ans := 0
	if n == 0 {
		return ans
	}
	m := map[byte]map[byte]int{}
	m['L'] = map[byte]int{
		'L': 0,
		'R': 0,
		'S': 0,
	}
	m['R'] = map[byte]int{
		'R': 0,
		'S': 1,
		'L': 2,
	}
	m['S'] = map[byte]int{
		'L': 1,
		'S': 0,
		'R': 0,
	}
	last := directions[0]
	r := 0
	if last == 'R' {
		r = 1
	}
	for i := 1; i < n; i++ {
		cur := directions[i]
		num := m[last][cur]
		ans += num
		if num > 0 && last == 'R' {
			ans += r - 1
		}
		if num > 0 {
			last = 'S'
		} else {
			last = cur
		}
		if cur != 'R' {
			r = 0
		} else {
			r++
		}

	}
	return ans
}
