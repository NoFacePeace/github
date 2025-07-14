package hash

func findLHS(nums []int) int {
	m := map[int]int{}
	for _, v := range nums {
		m[v]++
	}
	ans := 0
	for k, v := range m {
		if _, ok := m[k+1]; ok {
			ans = max(ans, m[k+1]+v)
		}
	}
	return ans
}
