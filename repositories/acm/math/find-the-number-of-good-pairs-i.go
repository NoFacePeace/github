package math

func numberOfPairs(nums1 []int, nums2 []int, k int) int {
	for i := range nums2 {
		nums2[i] *= k
	}
	cnt := 0
	for _, v := range nums1 {
		for _, vv := range nums2 {
			if vv == 0 {
				continue
			}
			if v%vv == 0 {
				cnt++
			}
		}
	}
	return cnt
}
