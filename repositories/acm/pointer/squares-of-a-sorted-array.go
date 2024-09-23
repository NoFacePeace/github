package pointer

func sortedSquares(nums []int) []int {
	n := len(nums)
	idx := -1
	for k, v := range nums {
		if v >= 0 {
			idx = k
			break
		}
		idx++
	}
	l := idx - 1
	r := idx
	if nums[idx] < 0 {
		l = n - 1
		r = n
	}
	arr := []int{}
	for l >= 0 && r < n {
		if -nums[l] > nums[r] {
			arr = append(arr, nums[r]*nums[r])
			r++
			continue
		}
		arr = append(arr, nums[l]*nums[l])
		l--
	}
	for l >= 0 {
		arr = append(arr, nums[l]*nums[l])
		l--
	}
	for r < n {
		arr = append(arr, nums[r]*nums[r])
		r++
	}
	return arr
}
