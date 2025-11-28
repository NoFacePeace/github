package hash

func getSneakyNumbers(nums []int) []int {
	m := map[int]bool{}
	ans := []int{}
	for _, v := range nums {
		if m[v] {
			ans = append(ans, v)
		}
		m[v] = true
	}
	return ans
}
