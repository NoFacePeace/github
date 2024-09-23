package hash

func twoSum(nums []int, target int) []int {
	m := map[int]int{}
	for k, v := range nums {
		sub := target - v
		if val, ok := m[sub]; ok {
			return []int{k, val}
		}
		m[v] = k
	}
	return nil
}
