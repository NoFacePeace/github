package array

func semiOrderedPermutation(nums []int) int {
	first := 0
	last := 0
	n := len(nums)
	for k := range nums {
		if nums[k] == 1 {
			first = k
		}
		if nums[k] == n {
			last = k
		}
	}
	if first < last {
		return first + n - last - 1
	}
	return first + n - last - 2
}
