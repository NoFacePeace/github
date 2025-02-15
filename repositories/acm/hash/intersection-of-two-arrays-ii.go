package hash

func intersect(nums1 []int, nums2 []int) []int {
	m := map[int]int{}
	for _, v := range nums1 {
		m[v]++
	}
	ans := []int{}
	for _, v := range nums2 {
		m[v]--
		if m[v] >= 0 {
			ans = append(ans, v)
		}
	}
	return ans
}
