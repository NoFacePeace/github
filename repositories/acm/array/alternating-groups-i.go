package array

func numberOfAlternatingGroups(colors []int) int {
	n := len(colors)
	if n < 3 {
		return 0
	}
	cnt := 0
	for i := 0; i < n; i++ {
		if colors[(i+n-1)%n] != colors[i] && colors[i] != colors[(i+n+1)%n] {
			cnt++
		}
	}
	return cnt
}
