package hash

func numRabbits(answers []int) int {
	m := map[int]int{}
	for _, v := range answers {
		m[v+1]++
	}
	ans := 0
	for k, v := range m {
		if k >= v {
			ans += k
			continue
		}
		ans += v / k * k
		if v/k*k != v {
			ans += k
		}
	}
	return ans
}
