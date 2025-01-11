package array

func occurrencesOfElement(nums []int, queries []int, x int) []int {
	idx := []int{}
	for i := range nums {
		if nums[i] == x {
			idx = append(idx, i)
		}
	}
	ans := []int{}
	for _, q := range queries {
		if q > len(idx) {
			ans = append(ans, -1)
		} else {
			ans = append(ans, idx[q-1])
		}
	}
	return ans
}
