package array

func minimumPairRemoval(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 0
	}
	if n == 2 {
		if nums[0] <= nums[1] {
			return 0
		}
		return 1
	}
	idx := 1
	mn := nums[0] + nums[1]
	ok := true
	for i := 1; i < n; i++ {
		if nums[i]+nums[i-1] < mn {
			mn = nums[i] + nums[i-1]
			idx = i
		}
		if nums[i] < nums[i-1] {
			ok = false
		}
	}
	if ok {
		return 0
	}
	arr := nums[:idx-1]
	arr = append(arr, mn)
	arr = append(arr, nums[idx+1:]...)
	return minimumPairRemoval(arr) + 1
}
