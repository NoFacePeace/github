package array

func containsNearbyDuplicate(nums []int, k int) bool {
	m := map[int]int{}
	for i, v := range nums {
		if val, ok := m[v]; ok {
			if i-val <= k {
				return true
			}
		}
		m[v] = i
	}
	return false
}
