package array

func avoidFlood(rains []int) []int {
	m := map[int]int{}
	n := len(rains)
	ans := make([]int, n)
	for k, v := range rains {
		if v == 0 {
			continue
		}
		if _, ok := m[v]; !ok {
			m[v] = k
			ans[k] = -1
			continue
		}
		ok := false
		for i := m[v] + 1; i < k; i++ {
			if rains[i] != 0 {
				continue
			}
			rains[i] = -1
			ans[i] = v
			ok = true
			m[v] = k
		}
		if !ok {
			return []int{}
		}
	}
	return ans
}
