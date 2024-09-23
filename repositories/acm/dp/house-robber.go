package dp

func rob(nums []int) int {
	n := len(nums)
	for i := 1; i < n; i++ {
		if i == 1 {
			nums[i] = max(nums[i], nums[i-1])
			continue
		}
		nums[i] = max(nums[i]+nums[i-2], nums[i-1])
	}
	return nums[n-1]
}
