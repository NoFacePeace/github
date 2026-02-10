package array

func constructTransformedArray(nums []int) []int {
	n := len(nums)
	result := make([]int, n)
	for i := 0; i < n; i++ {
		if nums[i] == 0 {
			continue
		}
		if nums[i] > 0 {
			idx := (i + nums[i]) % n
			result[i] = nums[idx]
			continue
		}
		idx := (i + nums[i]) % n
		if idx >= 0 {
			result[i] = nums[idx]
			continue
		}
		idx = -idx
		idx %= n
		idx = n - idx
		result[i] = nums[idx]
	}
	return result
}
