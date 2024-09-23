package array

func edgeScore(edges []int) int {
	m := map[int]int{}
	for k, v := range edges {
		m[v] += k
	}
	max := 0
	ans := 0
	for k, v := range m {
		if v > max {
			max = v
			ans = k
			continue
		}
		if v == max && k < ans {
			ans = k
		}
	}
	return ans
}
