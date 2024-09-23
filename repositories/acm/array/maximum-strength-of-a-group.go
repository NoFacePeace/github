package array

func maxStrength(nums []int) int64 {
	n := len(nums)
	if n == 0 {
		return 0
	}
	mx := nums[0]
	mn := nums[0]
	for i := 1; i < n; i++ {
		mx, mn = max(max(max(mx, nums[i]), mx*nums[i]), mn*nums[i]), min(min(min(mn, nums[i]), mn*nums[i]), mx*nums[i])
	}
	return int64(mx)
}
