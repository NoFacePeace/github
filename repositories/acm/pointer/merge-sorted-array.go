package pointer

func merge(nums1 []int, m int, nums2 []int, n int) {
	i := m - 1
	j := n - 1
	end := m + n - 1
	for i >= 0 && j >= 0 {
		if nums1[i] > nums2[j] {
			nums1[end] = nums1[i]
			i--
		} else {
			nums1[end] = nums2[j]
			j--
		}
		end--
	}
	for j >= 0 {
		nums1[end] = nums2[j]
		j--
		end--
	}
}
