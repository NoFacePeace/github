package pointer

func maxDistance(nums1 []int, nums2 []int) int {
	n := len(nums1)
	m := len(nums2)
	i := 0
	j := 0
	ans := 0
	for i < n && j < m {
		if i > j {
			j++
			continue
		}
		if nums1[i] > nums2[j] {
			i++
			continue
		}
		ans = max(ans, j-i)
		j++
	}
	return ans
}
