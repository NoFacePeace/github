package array

func sumOfBeauties(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	right := make([]int, n)
	mn := nums[n-1]
	for i := n - 2; i >= 0; i-- {
		right[i] = mn
		mn = min(mn, nums[i])
	}
	mx := nums[0]
	ans := 0
	for i := 1; i < n-1; i++ {
		if nums[i] > mx && nums[i] < right[i] {
			ans += 2
			mx = max(mx, nums[i])
			continue
		}
		if nums[i] > nums[i-1] && nums[i] < nums[i+1] {
			ans += 1
		}
		mx = max(mx, nums[i])
	}
	return ans
}
