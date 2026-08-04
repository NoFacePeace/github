package hash

func findMissingElements(nums []int) []int {
	n := len(nums)
	ans := []int{}
	if n == 0 {
		return ans
	}
	m := map[int]bool{}
	mn, mx := nums[0], nums[0]
	for _, num := range nums {
		mn = min(mn, num)
		mx = max(mx, num)
		m[num] = true
	}
	for i := mn; i <= mx; i++ {
		if _, ok := m[i]; !ok {
			ans = append(ans, i)
		}
	}
	return ans
}
