package hash

func countCompleteDayPairs(hours []int) int {
	for k := range hours {
		hours[k] %= 24
	}
	m := map[int]int{}
	cnt := 0
	for _, v := range hours {
		cnt += m[24-v]
		if v == 0 {
			cnt += m[0]
		}
		m[v]++
	}
	return cnt
}
