package hash

func containsNearbyDuplicate(nums []int, k int) bool {
	m := map[int]int{}
	for i, v := range nums {
		val, ok := m[v]
		if !ok {
			m[v] = i
			continue
		}
		if i-val <= k {
			return true
		}
		m[v] = i
	}
	return false
}
