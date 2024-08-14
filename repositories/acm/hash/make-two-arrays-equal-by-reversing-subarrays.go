package hash

func canBeEqual(target []int, arr []int) bool {
	m := map[int]int{}
	for _, v := range target {
		m[v]++
	}
	for _, v := range arr {
		if _, ok := m[v]; !ok {
			return false
		}
		m[v]--
		if m[v] == 0 {
			delete(m, v)
		}
	}
	if len(m) == 0 {
		return true
	}
	return false
}
